package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("catpc-secret-key-2024")

var db *sql.DB

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Stock       int     `json:"stock"`
	UserID      *int    `json:"user_id,omitempty"`
	Username    string  `json:"username,omitempty"`
	IsApproved  bool    `json:"is_approved"`
	CreatedAt   string  `json:"created_at,omitempty"`
}

type CartItem struct {
	ID        int     `json:"id"`
	ProductID int     `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Image     string  `json:"image"`
}

type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID int, username, role string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func GetUserFromToken(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, jwt.ErrSignatureInvalid
	}
	return ValidateJWT(tokenString)
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"error":   "Требуется авторизация",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := GetUserFromToken(tokenString)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"error":   "Неверный токен",
			})
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		return next(c)
	}
}

func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := c.Get("role").(string)
			hasRole := false

			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"success": false,
					"error":   "Недостаточно прав",
				})
			}

			return next(c)
		}
	}
}

func GetUserID(c echo.Context) int {
	return c.Get("user_id").(int)
}

func GetProducts(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT p.id, p.name, p.description, p.price, p.image, p.stock,
		       p.user_id, u.username, p.is_approved, p.created_at
		FROM products p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE p.is_approved = true
		ORDER BY p.id
		LIMIT $1 OFFSET $2
	`

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Ошибка загрузки товаров",
		})
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		var userID sql.NullInt64
		var username sql.NullString
		var createdAt sql.NullTime

		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Image, &p.Stock,
			&userID, &username, &p.IsApproved, &createdAt)
		if err != nil {
			continue
		}

		if userID.Valid {
			id := int(userID.Int64)
			p.UserID = &id
			p.Username = username.String
		}

		if createdAt.Valid {
			p.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}

		products = append(products, p)
	}

	var total int
	err = db.QueryRow("SELECT COUNT(*) FROM products WHERE is_approved = true").Scan(&total)
	if err != nil {
		total = len(products)
	}

	totalPages := 1
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"products":   products,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
			"total":      total,
		},
	})
}

func GetProductDetail(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверный ID",
		})
	}

	var product Product
	var userID sql.NullInt64
	var username sql.NullString
	var createdAt sql.NullTime

	err = db.QueryRow(`
		SELECT p.id, p.name, p.description, p.price, p.image, p.stock,
		       p.user_id, u.username, p.is_approved, p.created_at
		FROM products p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE p.id = $1
	`, id).Scan(&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Image, &product.Stock, &userID, &username, &product.IsApproved, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"success": false,
				"error":   "Товар не найден",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	if userID.Valid {
		id := int(userID.Int64)
		product.UserID = &id
		product.Username = username.String
	}

	if createdAt.Valid {
		product.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    product,
	})
}

func Register(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверные данные",
		})
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Все поля обязательны",
		})
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2",
		req.Username, req.Email).Scan(&count)

	if count > 0 {
		return c.JSON(http.StatusConflict, map[string]interface{}{
			"success": false,
			"error":   "Пользователь уже существует",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Ошибка хеширования пароля",
		})
	}

	var userID int
	err = db.QueryRow(`
		INSERT INTO users (username, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, 'customer', $4)
		RETURNING id
	`, req.Username, req.Email, string(hashedPassword), time.Now()).Scan(&userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	token, err := GenerateJWT(userID, req.Username, "customer")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Ошибка генерации токена",
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":       userID,
				"username": req.Username,
				"email":    req.Email,
				"role":     "customer",
			},
		},
	})
}

func Login(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверные данные",
		})
	}

	// Проверяем в базе данных
	var userID int
	var username, email, role, passwordHash string
	var isActive bool

	// Используем $1 два раза для поиска по username или email
	err := db.QueryRow(`
		SELECT id, username, email, role, password_hash, is_active
		FROM users WHERE username = $1 OR email = $1
	`, req.Username).Scan(&userID, &username, &email, &role, &passwordHash, &isActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"error":   "Неверный логин или пароль",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Ошибка базы данных",
		})
	}

	// Проверяем активность аккаунта
	if !isActive {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"error":   "Аккаунт заблокирован",
		})
	}

	// Сравниваем пароль
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "Неверный логин или пароль",
		})
	}

	// Генерируем JWT токен
	tokenString, err := GenerateJWT(userID, username, role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Ошибка генерации токена",
		})
	}

	// Формируем ответ
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"token": tokenString,
			"user": map[string]interface{}{
				"id":       userID,
				"username": username,
				"email":    email,
				"role":     role,
			},
		},
	})
}

func GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(int)

	var user User
	err := db.QueryRow(`
		SELECT id, username, email, role, is_active, created_at
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   "Пользователь не найден",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    user,
	})
}

func GetCart(c echo.Context) error {
	userID := GetUserID(c)

	rows, err := db.Query(`
		SELECT ci.id, ci.product_id, p.name, p.price, ci.quantity, p.image
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1 AND p.stock > 0 AND p.is_approved = true
		ORDER BY ci.added_at DESC
	`, userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	defer rows.Close()

	var cart []CartItem
	var total float64

	for rows.Next() {
		var item CartItem
		err := rows.Scan(&item.ID, &item.ProductID, &item.Name, &item.Price, &item.Quantity, &item.Image)
		if err != nil {
			continue
		}
		cart = append(cart, item)
		total += item.Price * float64(item.Quantity)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"items": cart,
			"total": total,
			"count": len(cart),
		},
	})
}

func AddToCart(c echo.Context) error {
	userID := GetUserID(c)

	var req struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверные данные",
		})
	}

	if req.Quantity <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Количество должно быть больше 0",
		})
	}

	var stock int
	var isApproved bool
	err := db.QueryRow(`
		SELECT stock, is_approved FROM products WHERE id = $1
	`, req.ProductID).Scan(&stock, &isApproved)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   "Товар не найден",
		})
	}

	if !isApproved {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Товар не доступен для покупки",
		})
	}

	if stock < req.Quantity {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Недостаточно товара в наличии",
		})
	}

	_, err = db.Exec(`
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
	`, userID, req.ProductID, req.Quantity)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Товар добавлен в корзину",
	})
}

func UpdateCartItem(c echo.Context) error {
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверный ID",
		})
	}

	userID := GetUserID(c)

	var req struct {
		Quantity int `json:"quantity"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверные данные",
		})
	}

	if req.Quantity <= 0 {
		_, err = db.Exec(`
			DELETE FROM cart_items
			WHERE id = $1 AND user_id = $2
		`, itemID, userID)
	} else {
		_, err = db.Exec(`
			UPDATE cart_items
			SET quantity = $1
			WHERE id = $2 AND user_id = $3
		`, req.Quantity, itemID, userID)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Корзина обновлена",
	})
}

func RemoveFromCart(c echo.Context) error {
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверный ID",
		})
	}

	userID := GetUserID(c)

	_, err = db.Exec(`
		DELETE FROM cart_items
		WHERE id = $1 AND user_id = $2
	`, itemID, userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Товар удален из корзины",
	})
}

func GetMyProducts(c echo.Context) error {
	userID := GetUserID(c)

	rows, err := db.Query(`
		SELECT id, name, description, price, image, stock, is_approved
		FROM products WHERE user_id = $1
		ORDER BY id
	`, userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Image, &p.Stock, &p.IsApproved); err == nil {
			products = append(products, p)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    products,
	})
}

func CreateProduct(c echo.Context) error {
	userID := GetUserID(c)
	role := c.Get("role").(string)

	var p Product
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверные данные",
		})
	}

	isApproved := role == "admin"

	var productID int
	err := db.QueryRow(`
		INSERT INTO products (name, description, price, image, stock, user_id, is_approved)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, p.Name, p.Description, p.Price, p.Image, p.Stock, userID, isApproved).Scan(&productID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	message := "Товар создан"
	if !isApproved {
		message += " (ожидает одобрения администратора)"
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": message,
		"data": map[string]interface{}{
			"id": productID,
		},
	})
}

func UpdateProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверный ID",
		})
	}

	userID := GetUserID(c)
	role := c.Get("role").(string)

	var ownerID int
	err = db.QueryRow("SELECT user_id FROM products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   "Товар не найден",
		})
	}

	if role != "admin" && ownerID != userID {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"error":   "Нет прав на редактирование",
		})
	}

	var p Product
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверные данные",
		})
	}

	if role != "admin" {
		_, err = db.Exec(`
			UPDATE products
			SET name = $1, description = $2, price = $3, image = $4, stock = $5, is_approved = false
			WHERE id = $6
		`, p.Name, p.Description, p.Price, p.Image, p.Stock, productID)
	} else {
		_, err = db.Exec(`
			UPDATE products
			SET name = $1, description = $2, price = $3, image = $4, stock = $5
			WHERE id = $6
		`, p.Name, p.Description, p.Price, p.Image, p.Stock, productID)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	message := "Товар обновлен"
	if role != "admin" {
		message += " (ожидает повторного одобрения)"
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": message,
	})
}

func DeleteProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверный ID",
		})
	}

	userID := GetUserID(c)
	role := c.Get("role").(string)

	var ownerID int
	err = db.QueryRow("SELECT user_id FROM products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   "Товар не найден",
		})
	}

	if role != "admin" && ownerID != userID {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"error":   "Нет прав на удаление",
		})
	}

	var inCart bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM cart_items WHERE product_id = $1)", productID).Scan(&inCart)

	if inCart {
		_, err = db.Exec("UPDATE products SET is_approved = false WHERE id = $1", productID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Товар скрыт (был в корзинах пользователей)",
		})
	}

	_, err = db.Exec("DELETE FROM products WHERE id = $1", productID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Товар удален",
	})
}

func GetAllUsers(c echo.Context) error {
	// Проверяем что пользователь администратор
	userRole := c.Get("role").(string)
	if userRole != "admin" {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"error":   "Недостаточно прав",
		})
	}

	// Проверяем существование поля is_protected в таблице
	var columnExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'users' AND column_name = 'is_protected'
		)
	`).Scan(&columnExists)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Ошибка проверки структуры БД",
		})
	}

	var query string
	if columnExists {
		// Если поле существует
		query = `
			SELECT id, username, email, role, is_active, is_protected, created_at
			FROM users ORDER BY created_at DESC
		`
	} else {
		// Если поле не существует (для обратной совместимости)
		query = `
			SELECT id, username, email, role, is_active, false as is_protected, created_at
			FROM users ORDER BY created_at DESC
		`
	}

	rows, err := db.Query(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	defer rows.Close()

	type UserDetail struct {
		ID          int    `json:"id"`
		Username    string `json:"username"`
		Email       string `json:"email"`
		Role        string `json:"role"`
		IsActive    bool   `json:"is_active"`
		IsProtected bool   `json:"is_protected"`
		CreatedAt   string `json:"created_at"`
	}

	var users []UserDetail
	for rows.Next() {
		var u UserDetail
		var createdAt time.Time
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.IsActive, &u.IsProtected, &createdAt)
		if err != nil {
			fmt.Printf("Ошибка сканирования пользователя: %v\n", err)
			continue
		}
		u.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		users = append(users, u)
	}

	if users == nil {
		users = []UserDetail{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    users,
	})
}

func UpdateUserRole(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверный ID",
		})
	}

	var req struct {
		Role string `json:"role"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверные данные",
		})
	}

	validRoles := map[string]bool{"customer": true, "seller": true, "admin": true}
	if !validRoles[req.Role] {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверная роль",
		})
	}

	// Получаем данные о пользователе, которого хотим изменить
	var targetUsername string
	var targetIsProtected bool
	err = db.QueryRow("SELECT username, is_protected FROM users WHERE id = $1", userID).
		Scan(&targetUsername, &targetIsProtected)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   "Пользователь не найден",
		})
	}

	// Проверяем, кто пытается изменить
	currentUserID := c.Get("user_id").(int)
	var currentUsername string
	var currentIsProtected bool
	err = db.QueryRow("SELECT username, is_protected FROM users WHERE id = $1", currentUserID).
		Scan(&currentUsername, &currentIsProtected)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Ошибка проверки прав",
		})
	}

	if targetIsProtected {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"error":   "Нельзя изменить роль защищенного пользователя",
		})
	}

	// Проверяем права на назначение роли администратора
	if req.Role == "admin" {
		// Только CatPC может назначать администраторов
		if currentUsername != "CatPC" {
			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"success": false,
				"error":   "Только главный администратор может назначать администраторов",
			})
		}
	}

	_, err = db.Exec("UPDATE users SET role = $1 WHERE id = $2", req.Role, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Роль пользователя обновлена",
	})
}

func ToggleUserActive(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверный ID",
		})
	}

	// Получаем данные о пользователе
	var targetUsername string
	var targetIsProtected bool
	var isActive bool

	err = db.QueryRow(`
		SELECT username, is_protected, is_active 
		FROM users WHERE id = $1
	`, userID).Scan(&targetUsername, &targetIsProtected, &isActive)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   "Пользователь не найден",
		})
	}

	// Запрещаем блокировку защищенного пользователя
	if targetIsProtected {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"error":   "Нельзя заблокировать защищенного пользователя",
		})
	}

	// Запрещаем блокировать самого себя
	currentUserID := c.Get("user_id").(int)
	if userID == currentUserID {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"error":   "Нельзя заблокировать себя",
		})
	}

	_, err = db.Exec("UPDATE users SET is_active = NOT is_active WHERE id = $1", userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	newStatus := "заблокирован"
	if !isActive {
		newStatus = "разблокирован"
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Пользователь " + newStatus,
	})
}

func GetPendingProducts(c echo.Context) error {
	rows, err := db.Query(`
		SELECT p.id, p.name, p.description, p.price, p.image, p.stock,
			   p.user_id, u.username, p.is_approved
		FROM products p
		JOIN users u ON p.user_id = u.id
		WHERE p.is_approved = false
		ORDER BY p.id
	`)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		var userID int
		var username string

		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Image, &p.Stock,
			&userID, &username, &p.IsApproved)
		if err != nil {
			continue
		}

		p.UserID = &userID
		p.Username = username
		products = append(products, p)
	}

	if products == nil {
		products = []Product{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    products,
	})
}

func ApproveProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверный ID",
		})
	}

	_, err = db.Exec("UPDATE products SET is_approved = true WHERE id = $1", productID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Товар одобрен",
	})
}

func ForceDeleteProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Неверный ID",
		})
	}

	_, err = db.Exec("DELETE FROM cart_items WHERE product_id = $1", productID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	_, err = db.Exec("DELETE FROM products WHERE id = $1", productID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Товар принудительно удален",
	})
}

func InitDB() error {
	connStr := "user=barsikuser password=barsik_password dbname=barsikdb host=localhost port=5432 sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ошибка ping БД: %v", err)
	}

	log.Println("✅ Успешно подключились к PostgreSQL!")
	return nil
}

func main() {
	if err := InitDB(); err != nil {
		panic(err)
	}
	defer db.Close()

	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: false,
		MaxAge:           3600,
	}))

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.POST("/api/register", Register)
	e.POST("/api/login", Login)
	e.GET("/api/products", GetProducts)
	e.GET("/api/products/:id", GetProductDetail)

	authGroup := e.Group("/api")
	authGroup.Use(AuthMiddleware)

	authGroup.GET("/profile", GetProfile)
	authGroup.GET("/cart", GetCart)
	authGroup.POST("/cart/add", AddToCart)
	authGroup.PUT("/cart/update/:id", UpdateCartItem)
	authGroup.DELETE("/cart/remove/:id", RemoveFromCart)

	sellerGroup := authGroup.Group("/seller")
	sellerGroup.Use(RequireRole("seller", "admin"))

	sellerGroup.GET("/my-products", GetMyProducts)
	sellerGroup.POST("/products", CreateProduct)
	sellerGroup.PUT("/products/:id", UpdateProduct)
	sellerGroup.DELETE("/products/:id", DeleteProduct)

	adminGroup := authGroup.Group("/admin")
	adminGroup.Use(RequireRole("admin"))

	adminGroup.GET("/users", GetAllUsers)
	adminGroup.PUT("/users/:id/role", UpdateUserRole)
	adminGroup.PUT("/users/:id/active", ToggleUserActive)
	adminGroup.GET("/pending-products", GetPendingProducts)
	adminGroup.PUT("/products/:id/approve", ApproveProduct)
	adminGroup.DELETE("/products/:id/force", ForceDeleteProduct)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "CatPC API работает! Используйте /api/ endpoints")
	})

	e.Logger.Fatal(e.Start(":1323"))
}

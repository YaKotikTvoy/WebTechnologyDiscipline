package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("secret-key"))

func init() {
    store.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   86400 * 7,
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
        Secure:   false,
    }
}

func initDB() error {
	connStr := "user=barsikuser password=barsik_password dbname=barsikdb host=localhost port=5432 sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ошибка ping БД: %v", err)
	}

	fmt.Println("✅ Успешно подключились к PostgreSQL!")
	return nil
}

func getSession(c echo.Context) (*sessions.Session, error) {
    //cookies := c.Request().Cookies()
    if sess, ok := c.Get("session").(*sessions.Session); ok {
        return sess, nil
    }


    //fmt.Println(cookies)
    sess, err := store.Get(c.Request(), "catpc-session")
    if err != nil {
        return nil, fmt.Errorf("сессия не найдена: %v", err)
    }


    c.Set("session", sess)
    return sess, nil
}

func getUserID(c echo.Context) (int, error) {
	sess, err := getSession(c)
	if err != nil {
		return 0, err
	}

	userIDValue := sess.Values["user_id"]
	if userIDValue == nil {
		return 0, fmt.Errorf("user_id не найден в сессии")
	}

	switch v := userIDValue.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("неверный формат user_id: %T", v)
	}
}

func requireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := getUserID(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Требуется авторизация",
			})
		}
		return next(c)
	}
}

func requireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := getSession(c)
			if err != nil {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Доступ запрещен",
				})
			}

			userRole, ok := sess.Values["role"].(string)
			if !ok {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Доступ запрещен",
				})
			}

			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Недостаточно прав",
				})
			}

			return next(c)
		}
	}
}

func main() {
	if err := initDB(); err != nil {
		panic(err)
	}
	defer db.Close()

	e := echo.New()

    e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
            return func(c echo.Context) error {
                fmt.Println("\n🔍 [Session Middleware] Начало для пути:", c.Path())

                // Получаем или создаем сессию
                sess, err := store.Get(c.Request(), "catpc-session")
                if err != nil {
                    fmt.Println("🆕 Создаем новую сессию")
                    sess, _ = store.New(c.Request(), "catpc-session")
                }

                // Настройки сессии
                sess.Options = &sessions.Options{
                    Path:     "/",
                    MaxAge:   86400 * 7,
                    HttpOnly: true,
                    SameSite: http.SameSiteLaxMode,
                    Secure:   false,
                }

                // Отладка: что в сессии сейчас
                fmt.Printf("📦 Сессия до обработки: user_id=%v, username=%v\n",
                    sess.Values["user_id"], sess.Values["username"])

                // Сохраняем в контекст
                c.Set("session", sess)

                // Выполняем следующий middleware/обработчик
                err = next(c)

                // Сохраняем сессию после обработки
                if err == nil && sess != nil {
                    if saveErr := sess.Save(c.Request(), c.Response()); saveErr != nil {
                        fmt.Printf("⚠️ Ошибка сохранения сессии: %v\n", saveErr)
                    } else {
                        fmt.Printf("💾 Сессия сохранена: user_id=%v\n", sess.Values["user_id"])
                    }
                }

                fmt.Println("🔍 [Session Middleware] Завершение\n")
                return err
            }
        })


        e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
            AllowOrigins:     []string{"http://localhost:5173"},
            AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
            AllowCredentials: true, // ← ВАЖНО: должно быть true
            ExposeHeaders:    []string{"Set-Cookie"},
            MaxAge:           3600,
        }))

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Всегда создаем/получаем сессию
            sess, err := store.Get(c.Request(), "catpc-session")
            if err != nil {
                // Если ошибка, создаем новую
                sess, _ = store.New(c.Request(), "catpc-session")
            }

            // Сохраняем в контекст
            c.Set("session", sess)

            // Выполняем запрос
            err = next(c)

            // Сохраняем сессию
            if err == nil && sess != nil {
                sess.Save(c.Request(), c.Response())
            }

            return err
        }
    })

	e.GET("/api/test-auth", testAuth)
	e.POST("/api/register", register)
	e.POST("/api/login", login)
	e.POST("/api/logout", logout)
	e.GET("/api/products", getProducts)
	e.GET("/api/products/:id", getProductDetail)
	e.GET("/api/set-test-cookie", func(c echo.Context) error {
        cookie := &http.Cookie{
            Name:     "test-cookie",
            Value:    "hello-world",
            Path:     "/",
            MaxAge:   3600,
            HttpOnly: true,
            SameSite: http.SameSiteLaxMode,
            Secure:   false,
        }
        c.SetCookie(cookie)

        return c.JSON(http.StatusOK, map[string]string{
            "message": "Test cookie set",
        })
    })



	authGroup := e.Group("/api")
	authGroup.Use(requireAuth)

	authGroup.GET("/cart", getCart)
	authGroup.POST("/cart/add", addToCart)
	authGroup.PUT("/cart/update/:id", updateCartItem)
	authGroup.DELETE("/cart/remove/:id", removeFromCart)

	authGroup.GET("/profile", getProfile)

	sellerGroup := authGroup.Group("/seller")
	sellerGroup.Use(requireRole("seller", "admin"))

	sellerGroup.GET("/my-products", getMyProducts)
	sellerGroup.POST("/products", createProduct)
	sellerGroup.PUT("/products/:id", updateProduct)
	sellerGroup.DELETE("/products/:id", deleteProduct)

	adminGroup := authGroup.Group("/admin")
	adminGroup.Use(requireRole("admin"))

	adminGroup.GET("/users", getAllUsers)
	adminGroup.PUT("/users/:id/role", updateUserRole)
	adminGroup.PUT("/users/:id/active", toggleUserActive)
	adminGroup.GET("/pending-products", getPendingProducts)
	adminGroup.PUT("/products/:id/approve", approveProduct)
	adminGroup.DELETE("/products/:id/force", forceDeleteProduct)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "CatPC API работает! Используйте /api/ endpoints")
	})

	fmt.Println("🚀 Сервер запущен на порту :1323")
	fmt.Println("🌐 CORS настроен для: http://localhost:5173")
	fmt.Println("🍪 Cookies разрешены")
	e.Logger.Fatal(e.Start(":1323"))
}

func register(c echo.Context) error {
	var req AuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2",
		req.Username, req.Email).Scan(&count)

	if count > 0 {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Пользователь уже существует"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка хеширования пароля"})
	}

	var userID int
	err = db.QueryRow(`
		INSERT INTO users (username, email, password_hash, role)
		VALUES ($1, $2, $3, 'customer')
		RETURNING id
	`, req.Username, req.Email, string(hashedPassword)).Scan(&userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	sess, err := getSession(c)
	if err != nil {
		sess, _ = store.New(c.Request(), "catpc-session")
		c.Set("session", sess)
	}

	sess.Values["user_id"] = userID
	sess.Values["username"] = req.Username
	sess.Values["role"] = "customer"

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Регистрация успешна",
		"user": map[string]interface{}{
			"id":       userID,
			"username": req.Username,
			"role":     "customer",
		},
	})
}

func login(c echo.Context) error {
	var req AuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
	}

	var user User
	var passwordHash string
	var isActive bool

	err := db.QueryRow(`
		SELECT id, username, email, role, password_hash, is_active
		FROM users WHERE username = $1 OR email = $1
	`, req.Username).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &passwordHash, &isActive)

	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный логин или пароль"})
	}

	if !isActive {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Аккаунт заблокирован"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный логин или пароль"})
	}


	fmt.Printf("\n✅ Аутентификация успешна: %s (ID: %d, Role: %s)\n",
		user.Username, user.ID, user.Role)

	cookies := c.Request().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "catpc-session" {
			c.SetCookie(&http.Cookie{
				Name:   "catpc-session",
				Value:  "",
				Path:   "/",
				MaxAge: -1,
				HttpOnly: true,
			})
		}
	}

	if oldSess, err := store.Get(c.Request(), "catpc-session"); err == nil {
		oldSess.Options.MaxAge = -1
		oldSess.Save(c.Request(), c.Response())
	}

	sess, err := store.New(c.Request(), "catpc-session")
	if err != nil {
		// Если не удалось создать новую, пробуем получить
		sess, _ = store.Get(c.Request(), "catpc-session")
	}

	for key := range sess.Values {
		delete(sess.Values, key)
	}

	sess.Values["user_id"] = user.ID
	sess.Values["username"] = user.Username
	sess.Values["role"] = user.Role
	sess.Values["email"] = user.Email

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	}

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		fmt.Printf("❌ КРИТИЧЕСКАЯ ОШИБКА сохранения сессии: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка создания сессии",
		})
	}

	c.Set("session", sess)

	fmt.Printf("✅ Сессия создана и сохранена: user_id=%v, username=%v, role=%v\n",
		sess.Values["user_id"], sess.Values["username"], sess.Values["role"])

	if checkSess, err := store.Get(c.Request(), "catpc-session"); err == nil {
		fmt.Printf("✅ Проверка сессии после сохранения: user_id=%v\n",
			checkSess.Values["user_id"])
	}


	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Авторизация успешна",
		"user":    user,
		"session_info": map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func logout(c echo.Context) error {
	sess, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"message": "Выход выполнен"})
	}

	sess.Options.MaxAge = -1
	delete(sess.Values, "user_id")
	delete(sess.Values, "username")
	delete(sess.Values, "role")

	return c.JSON(http.StatusOK, map[string]string{"message": "Выход выполнен"})
}

func getProductDetail(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}

	var product Product
	var userID sql.NullInt64
	var username sql.NullString

	err = db.QueryRow(`
		SELECT p.id, p.name, p.description, p.price, p.image, p.stock,
		       p.user_id, u.username, p.is_approved
		FROM products p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE p.id = $1
	`, id).Scan(&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Image, &product.Stock, &userID, &username, &product.IsApproved)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Товар не найден"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if userID.Valid {
		id := int(userID.Int64)
		product.UserID = &id
		product.Username = username.String
	}

	return c.JSON(http.StatusOK, product)
}

func getProfile(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	var user User
	err = db.QueryRow(`
		SELECT id, username, email, role
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Role)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
	}

	return c.JSON(http.StatusOK, user)
}

func getCart(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	rows, err := db.Query(`
		SELECT ci.id, ci.product_id, p.name, p.price, ci.quantity, p.image
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1 AND p.stock > 0 AND p.is_approved = true
		ORDER BY ci.added_at DESC
	`, userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer rows.Close()

	var cart []CartItem
	var total float64

	for rows.Next() {
		var item CartItem
		err := rows.Scan(&item.ID, &item.ProductID, &item.Name, &item.Price, &item.Quantity, &item.Image)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		cart = append(cart, item)
		total += item.Price * float64(item.Quantity)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": cart,
		"total": total,
		"count": len(cart),
	})
}

func addToCart(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	var req struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
	}

	var stock int
	var isApproved bool
	err = db.QueryRow(`
		SELECT stock, is_approved FROM products WHERE id = $1
	`, req.ProductID).Scan(&stock, &isApproved)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Товар не найден"})
	}

	if !isApproved {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Товар не доступен для покупки"})
	}

	if stock < req.Quantity {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Недостаточно товара в наличии. Доступно: %d шт.", stock),
		})
	}

	_, err = db.Exec(`
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
	`, userID, req.ProductID, req.Quantity)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Товар добавлен в корзину"})
}

func updateCartItem(c echo.Context) error {
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}

	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	var req struct {
		Quantity int `json:"quantity"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
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
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Корзина обновлена"})
}

func removeFromCart(c echo.Context) error {
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}

	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	_, err = db.Exec(`
		DELETE FROM cart_items
		WHERE id = $1 AND user_id = $2
	`, itemID, userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Товар удален из корзины"})
}

func getProducts(c echo.Context) error {
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	query := `
		SELECT p.id, p.name, p.description, p.price, p.image, p.stock,
		       p.user_id, u.username, p.is_approved
		FROM products p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE p.is_approved = true
		ORDER BY p.id
		LIMIT $1 OFFSET $2
	`

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка загрузки товаров",
		})
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		var userID sql.NullInt64
		var username sql.NullString

		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Image, &p.Stock,
			&userID, &username, &p.IsApproved)
		if err != nil {
			continue
		}

		if userID.Valid {
			id := int(userID.Int64)
			p.UserID = &id
			p.Username = username.String
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
		"products":   products,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
		"total":      total,
	})
}



func getMyProducts(c echo.Context) error {
    userID, err := getUserID(c)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
    }

    // Всегда инициализируем пустой массив
    products := []Product{}

    rows, err := db.Query(`
        SELECT id, name, description, price, image, stock, is_approved
        FROM products WHERE user_id = $1
        ORDER BY id
    `, userID)

    if err != nil {
        // Просто логируем и возвращаем пустой массив
        fmt.Printf("⚠️ Ошибка запроса товаров: %v\n", err)
    } else {
        defer rows.Close()

        for rows.Next() {
            var p Product
            if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Image, &p.Stock, &p.IsApproved); err == nil {
                products = append(products, p)
            }
        }
    }

    return c.JSON(http.StatusOK, products)
}
/*func getMyProducts(c echo.Context) error {
    fmt.Println("🔍 Вызван getMyProducts")

    userID, err := getUserID(c)
    if err != nil {
        fmt.Println("❌ Ошибка getUserID:", err)
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
    }

    fmt.Printf("✅ UserID: %d\n", userID)

    rows, err := db.Query(`
        SELECT id, name, description, price, image, stock, is_approved
        FROM products WHERE user_id = $1
        ORDER BY id
    `, userID)

    if err != nil {
        fmt.Printf("❌ Ошибка запроса: %v\n", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Image, &p.Stock, &p.IsApproved)
        if err != nil {
            fmt.Printf("❌ Ошибка scan: %v\n", err)
            continue
        }
        products = append(products, p)
    }

    fmt.Printf("✅ Найдено товаров: %d\n", len(products))

    return c.JSON(http.StatusOK, products)
}
*/

func createProduct(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	sess, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	var p Product
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
	}

	var isApproved bool
	userRole, _ := sess.Values["role"].(string)
	if userRole == "admin" {
		isApproved = true
	} else {
		isApproved = false
	}

	var productID int
	err = db.QueryRow(`
		INSERT INTO products (name, description, price, image, stock, user_id, is_approved)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, p.Name, p.Description, p.Price, p.Image, p.Stock, userID, isApproved).Scan(&productID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	message := "Товар создан"
	if !isApproved {
		message += " (ожидает одобрения администратора)"
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": message,
		"id":      productID,
	})
}

func updateProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}

	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	sess, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	var ownerID int
	err = db.QueryRow("SELECT user_id FROM products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Товар не найден"})
	}

	userRole, _ := sess.Values["role"].(string)
	if userRole != "admin" && ownerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Нет прав на редактирование"})
	}

	var p Product
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
	}

	if userRole != "admin" {
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
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	message := "Товар обновлен"
	if userRole != "admin" {
		message += " (ожидает повторного одобрения)"
	}

	return c.JSON(http.StatusOK, map[string]string{"message": message})
}

func deleteProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}

	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	sess, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Требуется авторизация"})
	}

	var ownerID int
	err = db.QueryRow("SELECT user_id FROM products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Товар не найден"})
	}

	userRole, _ := sess.Values["role"].(string)
	if userRole != "admin" && ownerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Нет прав на удаление"})
	}

	var inCart bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM cart_items WHERE product_id = $1)", productID).Scan(&inCart)

	if inCart {
		_, err = db.Exec("UPDATE products SET is_approved = false WHERE id = $1", productID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Товар скрыт (был в корзинах пользователей)"})
	}

	_, err = db.Exec("DELETE FROM products WHERE id = $1", productID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Товар удален"})
}

func getAllUsers(c echo.Context) error {
	rows, err := db.Query(`
		SELECT id, username, email, role, is_active, created_at
		FROM users ORDER BY created_at DESC
	`)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer rows.Close()

	type UserDetail struct {
		ID        int       `json:"id"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		Role      string    `json:"role"`
		IsActive  bool      `json:"is_active"`
		CreatedAt time.Time `json:"created_at"`
	}

	var users []UserDetail
	for rows.Next() {
		var u UserDetail
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		users = append(users, u)
	}

	return c.JSON(http.StatusOK, users)
}

func updateUserRole(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}

	var req struct {
		Role string `json:"role"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
	}

	validRoles := map[string]bool{"customer": true, "seller": true, "admin": true}
	if !validRoles[req.Role] {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверная роль"})
	}

	_, err = db.Exec("UPDATE users SET role = $1 WHERE id = $2", req.Role, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Роль пользователя обновлена"})
}

func toggleUserActive(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}

	var isActive bool
	err = db.QueryRow("SELECT is_active FROM users WHERE id = $1", userID).Scan(&isActive)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
	}

	_, err = db.Exec("UPDATE users SET is_active = NOT is_active WHERE id = $1", userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	newStatus := "заблокирован"
	if !isActive {
		newStatus = "разблокирован"
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Пользователь %s", newStatus),
	})
}

func getPendingProducts(c echo.Context) error {
    rows, err := db.Query(`
        SELECT p.id, p.name, p.description, p.price, p.image, p.stock,
               p.user_id, u.username, p.is_approved
        FROM products p
        JOIN users u ON p.user_id = u.id
        WHERE p.is_approved = false
        ORDER BY p.id
    `)

    if err != nil {
        // В случае ошибки возвращаем пустой массив
        return c.JSON(http.StatusOK, []Product{})
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

    // ВАЖНО: Если товаров нет, возвращаем пустой массив, а не null
    if products == nil {
        products = []Product{}
    }

    return c.JSON(http.StatusOK, products)
}

func approveProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}

	_, err = db.Exec("UPDATE products SET is_approved = true WHERE id = $1", productID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Товар одобрен"})
}

func forceDeleteProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}

	_, err = db.Exec("DELETE FROM cart_items WHERE product_id = $1", productID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	_, err = db.Exec("DELETE FROM products WHERE id = $1", productID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Товар принудительно удален"})
}

func testAuth(c echo.Context) error {
	sess, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"authenticated": false,
			"message":       "Сессия не найдена",
		})
	}

	userIDValue := sess.Values["user_id"]
	username, _ := sess.Values["username"].(string)

	var userID int
	if userIDValue != nil {
		switch v := userIDValue.(type) {
		case int:
			userID = v
		case int64:
			userID = int(v)
		case float64:
			userID = int(v)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"authenticated": userID > 0,
		"user_id":       userID,
		"username":      username,
		"cookies":       c.Request().Cookies(),
	})
}
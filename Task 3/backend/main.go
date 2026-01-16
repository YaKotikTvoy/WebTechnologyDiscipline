package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

var db *sql.DB

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

func main() {
	if err := initDB(); err != nil {
		panic(err)
	}
	defer db.Close()

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.RequestLogger())


	e.GET("/api/products", getProducts)
	e.GET("/api/products/stats", getProductsStats)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "CatPC Backend работает! База: barsikdb")
	})
    e.GET("/api/products/:id", getProduct)

	fmt.Println("🚀 Сервер запущен на порту :1323")
	e.Logger.Fatal(e.Start(":1323"))
}

func getProducts(c echo.Context) error {

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	search := c.QueryParam("search")
	sortBy := c.QueryParam("sort")
	availability := c.QueryParam("availability")

	query := "SELECT id, name, description, price, image, stock, created_at FROM products"
	countQuery := "SELECT COUNT(*) FROM products"

	var conditions []string
	var args []interface{}
	argIndex := 1


	if search != "" {
		conditions = append(conditions,
			fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}


	if availability != "" {
		switch availability {
		case "in_stock":
			conditions = append(conditions, fmt.Sprintf("stock > 0"))
		case "out_of_stock":
			conditions = append(conditions, fmt.Sprintf("stock = 0"))
		case "low_stock":
			conditions = append(conditions, fmt.Sprintf("stock > 0 AND stock < 3"))
		}
	}


	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		query += whereClause
		countQuery += whereClause
	}


	switch sortBy {
	case "price_asc":
		query += " ORDER BY price ASC"
	case "price_desc":
		query += " ORDER BY price DESC"
	case "name":
		query += " ORDER BY name"
	default:
		query += " ORDER BY id"
	}


	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)


	rows, err := db.Query(query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Ошибка запроса: %v", err),
		})
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Image, &p.Stock, &p.CreatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Ошибка чтения данных: %v", err),
			})
		}
		products = append(products, p)
	}


	var total int
	countArgs := args[:len(args)-2]
	if len(countArgs) > 0 {
		err = db.QueryRow(countQuery, countArgs...).Scan(&total)
	} else {
		err = db.QueryRow(countQuery).Scan(&total)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Ошибка подсчета: %v", err),
		})
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(http.StatusOK, map[string]interface{}{
		"products":   products,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
		"total":      total,
	})
}

func getProductsStats(c echo.Context) error {
	var stats Stats


	err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&stats.Total)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}


	err = db.QueryRow("SELECT COUNT(*) FROM products WHERE stock > 0").Scan(&stats.InStock)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	err = db.QueryRow("SELECT COUNT(*) FROM products WHERE stock > 0 AND stock < 3").Scan(&stats.LowStock)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	stats.OutOfStock = stats.Total - stats.InStock

	return c.JSON(http.StatusOK, stats)
}

func getProduct(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный ID"})
	}

	var product Product
	err = db.QueryRow(`
		SELECT id, name, description, price, image, stock, created_at
		FROM products WHERE id = $1
	`, id).Scan(&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Image, &product.Stock, &product.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Товар не найден"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, product)
}
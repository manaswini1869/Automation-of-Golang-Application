package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus metrics
var (
	addGoalCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "add_goal_requests_total",
		Help: "Total number of add goal requests",
	})
	removeGoalCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "remove_goal_requests_total",
		Help: "Total number of remove goal requests",
	})
	httpRequestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
		[]string{"path"},
	)
)

func init() {
	prometheus.MustRegister(addGoalCounter)
	prometheus.MustRegister(removeGoalCounter)
	prometheus.MustRegister(httpRequestsCounter)
}

func createConnection() (*sql.DB, error) {
	connectionStr := fmt.Sprintf("user=%s password=%s host=%s dbName=%s SSLmode=%s",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_DBNAME"),
		os.Getenv("SSL"),
	)
	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob(os.Getenv("KO_DATA_PATH") + "/*")
	db, err := createConnection()
	if err != nil {
		log.Println("Error connecting to PostgreSQL", err)
		return
	}
	defer db.Close()
	router.GET("/goals", func(c *gin.Context) {
		rows, err := db.Query("SELECT * FROM goals")
		if err != nil {
			log.Println("Error from querying database", err)
			c.String(http.StatusInternalServerError, "Error querying the database")
			return
		}
		defer rows.Close()

		var goals []struct {
			ID   int
			Name string
		}

		for rows.Next() {
			var goal struct {
				ID   int
				Name string
			}
			if err := rows.Scan(&goal.ID, &goal.Name); err != nil {
				log.Println("Error scanning rows", err)
				continue
			}
			goals = append(goals, goal)
		}
		httpRequestsCounter.WithLabelValues("/").Inc()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"goals": goals,
		})
	})

	router.POST("/add_goal", func(c *gin.Context) {

	})
}

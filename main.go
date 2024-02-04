package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	//"github.com/bxcodec/faker/v3"
	//"github.com/go-faker/faker/v4"
	"github.com/ddosify/go-faker/faker"
	"github.com/icrowley/fake"

	"github.com/joho/godotenv"
)

type Review struct {
	User    string `json:"user"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

type Product struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Price        float64  `json:"price"`
	Brand        string   `json:"brand"`
	Category     string   `json:"category"`
	Color        string   `json:"color"`
	Size         string   `json:"size"`
	Availability int      `json:"availability"`
	Reviews      []Review `json:"reviews"`
}

func main() {

	// Set client options and connect to MongoDB
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("error loading .env file")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.TODO(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	// List databases
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Database: %v\n", databases)

	// Create JSON files
	createJSON()
	// Insert the JSON files into MongoDB
	InsertJsonIntoDB(client)

}

func createJSON() {
	// Initialize the faker library
	//faker, err := faker.New("en")
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate 10 fake products
	for i := 0; i < 10; i++ {
		product := Product{
			ID:           fmt.Sprintf("product%d", i+1),
			Name:         faker.NewFaker().RandomProduct(),
			Description:  fake.Sentence(),
			Price:        rand.Float64() * 100,
			Brand:        fake.Brand(),
			Category:     fake.Word(),
			Color:        fake.Color(),
			Size:         fake.Word(),
			Availability: rand.Intn(100),
			Reviews:      make([]Review, rand.Intn(5)),
		}

		// Generate reviews for the product
		for j := 0; j < len(product.Reviews); j++ {
			review := Review{
				User:    fake.MaleFullName(),
				Rating:  rand.Intn(5) + 1,
				Comment: fake.ParagraphsN(2),
			}
			product.Reviews[j] = review
		}

		// Save the product as a JSON file
		data, err := json.MarshalIndent(product, "", "  ")
		if err != nil {
			panic(err)
		}
		filename := fmt.Sprintf("product%d.json", i+1)
		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Saved product %d as %s\n", i+1, filename)
	}
}

// create a connection to mongodb
func InsertJsonIntoDB(client *mongo.Client) {
	collection := client.Database("test").Collection("products")
	// Create a context for the operation
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	// Read the JSON files and insert them into MongoDB
	for i := 0; i < 10; i++ {
		// Read the JSON file
		data, err := os.ReadFile(fmt.Sprintf("product%d.json", i+1))
		if err != nil {
			log.Fatal(err)
		}
		// Unmarshal the JSON data into a Product struct
		var product Product
		err = json.Unmarshal(data, &product)
		if err != nil {
			log.Fatal(err)
		}
		// Insert the product into the collection
		_, err = collection.InsertOne(ctx, product)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Inserted product %d into the collection\n", i+1)
	}
}
package main

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// Dynamic SQL

const dsn = "root:yotoo123qwe@tcp(192.168.31.120:3306)/yoo?charset=utf8mb4&parseTime=True&loc=Local"

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "internal/pkg/model",
	})
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect database: %v", err)
		return
	}

	g.UseDB(db)

	g.GenerateModelAs("users", "UserM")
	g.GenerateModelAs("plans", "PlanM")
	g.GenerateModelAs("projects", "ProjectM", gen.FieldType("tags", "datatypes.JSON"))

	g.Execute()
}

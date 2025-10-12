package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	daytona "github.com/PhilippBuschhaus/daytona-sdk-go"
)

func main() {
	fmt.Println("=== PySpark with Declarative Builder Example ===\n")

	// Create SDK client
	// API key is loaded from DAYTONA_API_KEY environment variable
	client, err := daytona.NewClient(&daytona.Config{})
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Build a custom image with PySpark pre-installed using declarative builder
	// Following the pattern from the TypeScript SDK's Image class
	image := daytona.DebianSlim("3.11").
		AptInstall([]string{"openjdk-17-jre-headless", "procps"}).
		EnvVars(
			"JAVA_HOME", "/usr/lib/jvm/java-17-openjdk-amd64",
			"PYSPARK_PYTHON", "python3",
			"SPARK_LOCAL_IP", "127.0.0.1",
			"PYTHONUNBUFFERED", "1",
		).
		PipInstall([]string{
			"pyspark==3.5.0",
			"pandas",
			"numpy",
			"pyarrow",
			"matplotlib",
			"seaborn",
		}).
		Workdir("/workspace")

	fmt.Println("Creating sandbox with pre-installed PySpark...")
	fmt.Println("This will build a custom image with all dependencies pre-installed.")
	fmt.Println("Initial build may take a few minutes, but subsequent uses will be faster.\n")

	createReq := &daytona.CreateSandboxRequest{
		Target:            daytona.StringPtr("eu"),          // Required: deployment region
		DockerfileContent: daytona.StringPtr(image.Build()), // Custom image with PySpark

		// Optional: Labels for organization
		Labels: map[string]string{
			"example": "pyspark",
			"stack":   "pyspark",
		},
	}

	sandbox, err := client.CreateSandbox(ctx, createReq)
	if err != nil {
		log.Fatal("Failed to create sandbox:", err)
	}
	sandboxID := sandbox.GetId()
	fmt.Printf("✓ Created sandbox: %s\n", sandboxID)

	// Wait for sandbox to be ready (longer timeout for image build)
	fmt.Println("\nWaiting for sandbox to be ready (building custom image)...")
	sandbox, err = client.WaitForSandboxReady(ctx, sandboxID, 10*time.Minute)
	if err != nil {
		log.Fatal("Failed to wait for sandbox:", err)
	}
	fmt.Println("✓ Sandbox with pre-installed PySpark is ready!")

	// Verify PySpark is pre-installed and ready to use
	fmt.Println("\nVerifying PySpark installation...")
	verifyReq := &daytona.ExecuteCommandRequest{
		Command: `python3 -c "from pyspark.sql import SparkSession; spark = SparkSession.builder.appName('test').getOrCreate(); print(f'✓ PySpark {spark.version} is ready!'); spark.stop()"`,
		Timeout: 30.0,
	}

	verifyResp, err := client.ExecuteCommand(ctx, sandboxID, verifyReq)
	if err != nil {
		log.Printf("Failed to verify PySpark: %v\n", err)
	} else {
		fmt.Println(verifyResp.Result)
	}

	// Create and run a PySpark analysis script
	fmt.Println("\nCreating PySpark analysis script...")
	pysparkScript := generatePySparkScript()

	err = client.WriteFile(ctx, sandboxID, "/tmp/analysis.py", []byte(pysparkScript))
	if err != nil {
		log.Fatal("Failed to write PySpark script:", err)
	}
	fmt.Println("✓ Analysis script created")

	// Run the PySpark analysis - no installation needed, just import and use!
	fmt.Println("\nRunning PySpark analysis (using pre-installed packages)...")
	execReq := &daytona.ExecuteCommandRequest{
		Command: "python3 /tmp/analysis.py 2>&1",
		Cwd:     "/tmp",
		Timeout: 120.0, // Give more time for Spark operations
	}

	execResp, err := client.ExecuteCommand(ctx, sandboxID, execReq)
	if err != nil {
		log.Fatal("Failed to execute PySpark script:", err)
	}

	fmt.Println("\n=== PySpark Analysis Output ===")
	fmt.Println(execResp.Result)

	// Read generated files
	fmt.Println("\n=== Generated Files ===")

	// Read sales summary
	summaryData, err := client.ReadFile(ctx, sandboxID, "/tmp/sales_summary.csv")
	if err != nil {
		log.Printf("Failed to read sales summary: %v\n", err)
	} else {
		fmt.Println("\nSales Summary (CSV):")
		fmt.Println(string(summaryData))
	}

	// List all generated files
	fmt.Println("\nGenerated files in /tmp:")
	files, err := client.ListFiles(ctx, sandboxID, "/tmp")
	if err != nil {
		log.Printf("Failed to list files: %v\n", err)
	} else {
		for _, f := range files {
			if strings.HasSuffix(f.Name, ".csv") || strings.HasSuffix(f.Name, ".parquet") {
				fmt.Printf("  - %s [%.0f bytes]\n", f.Name, f.Size)
			}
		}
	}

	// Clean up
	fmt.Println("\nCleaning up...")
	err = client.DeleteSandbox(ctx, sandboxID)
	if err != nil {
		log.Printf("Failed to delete sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox deleted")
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("Successfully used declarative builder to create a PySpark environment!")
	fmt.Println("All packages were pre-installed and ready to import immediately.")
}

// generatePySparkScript creates a PySpark analysis that can immediately import and use PySpark
func generatePySparkScript() string {
	return `#!/usr/bin/env python3
# PySpark is pre-installed via declarative builder - just import and use!
from pyspark.sql import SparkSession
from pyspark.sql.functions import col, sum as spark_sum, avg, count, max as spark_max, min as spark_min
import pandas as pd
import random
from datetime import datetime, timedelta

print("Starting PySpark analysis with pre-installed packages...", flush=True)

# Create Spark session - PySpark is already installed
spark = SparkSession.builder \
    .appName("DeclarativeBuilderExample") \
    .config("spark.master", "local[1]") \
    .config("spark.driver.memory", "512m") \
    .config("spark.sql.shuffle.partitions", "2") \
    .getOrCreate()

spark.sparkContext.setLogLevel("ERROR")
print(f"✓ Using pre-installed PySpark {spark.version}", flush=True)

# Generate sample e-commerce data
print("\nGenerating sample data...", flush=True)
num_records = 500  # Reduced for faster execution
data = []

products = ['Laptop', 'Phone', 'Tablet', 'Monitor', 'Keyboard', 'Mouse', 'Headphones']
categories = ['Electronics', 'Computers', 'Accessories'] 
regions = ['North America', 'Europe', 'Asia', 'South America']

start_date = datetime(2024, 1, 1)
for i in range(num_records):
    transaction = {
        'transaction_id': f'TXN-{i+1:06d}',
        'date': (start_date + timedelta(days=random.randint(0, 365))).strftime('%Y-%m-%d'),
        'product': random.choice(products),
        'category': random.choice(categories),
        'quantity': random.randint(1, 10),
        'unit_price': round(random.uniform(10, 2000), 2),
        'region': random.choice(regions),
        'customer_id': f'CUST-{random.randint(1000, 5000):04d}',
    }
    data.append(transaction)

# Create DataFrame
df = spark.createDataFrame(data)
df = df.withColumn("total_amount", col("quantity") * col("unit_price"))

print(f"✓ Generated {df.count()} transactions", flush=True)

# Perform analytics using pre-installed PySpark
print("\n=== ANALYSIS RESULTS ===", flush=True)

# 1. Sales by Region
print("\n1. Sales by Region:", flush=True)
sales_by_region = df.groupBy("region") \
    .agg(
        spark_sum("total_amount").alias("total_sales"),
        count("transaction_id").alias("num_transactions"),
        avg("total_amount").alias("avg_transaction")
    ) \
    .orderBy("total_sales", ascending=False)

sales_by_region.show(truncate=False)

# 2. Top Products
print("\n2. Top 5 Products by Revenue:", flush=True)
top_products = df.groupBy("product") \
    .agg(
        spark_sum("total_amount").alias("revenue"),
        spark_sum("quantity").alias("units_sold")
    ) \
    .orderBy("revenue", ascending=False) \
    .limit(5)

top_products.show(truncate=False)

# Save results using pre-installed pandas
print("\n=== SAVING RESULTS ===", flush=True)

# Convert to Pandas (pre-installed via declarative builder)
sales_summary = sales_by_region.toPandas()
sales_summary.to_csv('/tmp/sales_summary.csv', index=False)
print("✓ Saved sales summary to /tmp/sales_summary.csv", flush=True)

top_products_pd = top_products.toPandas()
top_products_pd.to_csv('/tmp/top_products.csv', index=False)
print("✓ Saved top products to /tmp/top_products.csv", flush=True)

# Summary statistics
total_revenue = df.agg(spark_sum("total_amount")).collect()[0][0]
total_orders = df.count()
unique_customers = df.select("customer_id").distinct().count()

print(f"\n=== SUMMARY ===", flush=True)
print(f"Total Revenue: ${total_revenue:,.2f}", flush=True)
print(f"Total Orders: {total_orders}", flush=True)
print(f"Unique Customers: {unique_customers}", flush=True)
print(f"Average Order Value: ${total_revenue/total_orders:,.2f}", flush=True)

spark.stop()
print("\n✓ Analysis complete - all packages were pre-installed via declarative builder!", flush=True)
`
}

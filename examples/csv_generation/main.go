package main

import (
	"context"
	"fmt"
	"log"
	"time"

	daytona "github.com/PhilippBuschhaus/daytona-sdk-go"
)

func main() {
	fmt.Println("=== CSV Data Generation Example ===\n")

	// Create SDK client
	// API key is loaded from DAYTONA_API_KEY environment variable
	client, err := daytona.NewClient(&daytona.Config{})
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Create a new sandbox
	fmt.Println("Creating sandbox...")
	createReq := &daytona.CreateSandboxRequest{
		Target: daytona.StringPtr("eu"), // Required: deployment region

		// Optional: Add labels for organization
		Labels: map[string]string{
			"example": "csv_generation",
		},
	}

	sandbox, err := client.CreateSandbox(ctx, createReq)
	if err != nil {
		log.Fatal("Failed to create sandbox:", err)
	}
	sandboxID := sandbox.GetId()
	fmt.Printf("✓ Created sandbox: %s\n", sandboxID)

	// Wait for sandbox to be ready
	fmt.Println("\nWaiting for sandbox to be ready...")
	sandbox, err = client.WaitForSandboxReady(ctx, sandboxID, 5*time.Minute)
	if err != nil {
		log.Fatal("Failed to wait for sandbox:", err)
	}
	fmt.Println("✓ Sandbox is ready!")

	// Create Python script for generating CSV with mock data
	fmt.Println("\nCreating Python script for CSV generation...")
	pythonScript := `#!/usr/bin/env python3
import csv
import random
from datetime import datetime, timedelta

# Generate mock sales data
def generate_sales_data(num_records=100):
    products = ['Laptop', 'Phone', 'Tablet', 'Monitor', 'Keyboard', 'Mouse', 'Headphones', 'Webcam']
    regions = ['North', 'South', 'East', 'West', 'Central']
    
    data = []
    start_date = datetime(2024, 1, 1)
    
    for i in range(num_records):
        record = {
            'order_id': f'ORD-{i+1:05d}',
            'date': (start_date + timedelta(days=random.randint(0, 365))).strftime('%Y-%m-%d'),
            'product': random.choice(products),
            'quantity': random.randint(1, 10),
            'unit_price': round(random.uniform(50, 2000), 2),
            'region': random.choice(regions),
            'customer_id': f'CUST-{random.randint(1000, 9999)}',
            'discount': round(random.uniform(0, 0.3), 2)
        }
        record['total_price'] = round(record['quantity'] * record['unit_price'] * (1 - record['discount']), 2)
        data.append(record)
    
    return data

# Write data to CSV
def write_to_csv(data, filename='sales_data.csv'):
    if not data:
        return
    
    with open(filename, 'w', newline='') as csvfile:
        fieldnames = data[0].keys()
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        
        writer.writeheader()
        writer.writerows(data)
    
    print(f"Generated {len(data)} records in {filename}")
    return filename

# Generate statistics
def generate_stats(data):
    total_sales = sum(record['total_price'] for record in data)
    avg_order = total_sales / len(data) if data else 0
    
    product_sales = {}
    for record in data:
        product = record['product']
        if product not in product_sales:
            product_sales[product] = 0
        product_sales[product] += record['total_price']
    
    top_product = max(product_sales.items(), key=lambda x: x[1]) if product_sales else ('N/A', 0)
    
    print(f"\n=== Sales Statistics ===")
    print(f"Total Sales: ${total_sales:,.2f}")
    print(f"Average Order Value: ${avg_order:,.2f}")
    print(f"Top Product: {top_product[0]} (${top_product[1]:,.2f})")
    print(f"Total Orders: {len(data)}")

if __name__ == "__main__":
    print("Generating mock sales data...")
    sales_data = generate_sales_data(100)
    
    csv_file = write_to_csv(sales_data)
    generate_stats(sales_data)
    
    print(f"\n✓ CSV file '{csv_file}' has been created successfully!")
`

	err = client.WriteFile(ctx, sandboxID, "/tmp/generate_csv.py", []byte(pythonScript))
	if err != nil {
		log.Fatal("Failed to write Python script:", err)
	}
	fmt.Println("✓ Python script created")

	// Execute the Python script to generate CSV
	fmt.Println("\nExecuting Python script to generate CSV...")
	execReq := &daytona.ExecuteCommandRequest{
		Command: "python3 /tmp/generate_csv.py",
		Cwd:     "/tmp",
		Timeout: 30.0,
	}

	execResp, err := client.ExecuteCommand(ctx, sandboxID, execReq)
	if err != nil {
		log.Fatal("Failed to execute Python script:", err)
	}
	if execResp.ExitCode != 0 {
		log.Fatalf("Script failed with exit code %.0f: %s", execResp.ExitCode, execResp.Result)
	}
	fmt.Println("\nScript output:")
	fmt.Println(execResp.Result)

	// Read the generated CSV file
	fmt.Println("\nReading generated CSV file...")
	csvData, err := client.ReadFile(ctx, sandboxID, "/tmp/sales_data.csv")
	if err != nil {
		log.Fatal("Failed to read CSV file:", err)
	}

	// Display first few lines of CSV
	csvString := string(csvData)
	lines := 0
	maxLines := 10
	fmt.Println("\nFirst 10 rows of generated CSV:")
	fmt.Println("=" + string(make([]byte, 80, 80)))
	for i, char := range csvString {
		fmt.Print(string(char))
		if char == '\n' {
			lines++
			if lines >= maxLines {
				fmt.Println("...")
				fmt.Printf("[%d more bytes of data]\n", len(csvData)-i)
				break
			}
		}
	}
	fmt.Println("=" + string(make([]byte, 80, 80)))

	// Get file info
	fmt.Println("\nGetting CSV file information...")
	fileInfo, err := client.GetFileInfo(ctx, sandboxID, "/tmp/sales_data.csv")
	if err != nil {
		log.Printf("Failed to get file info: %v\n", err)
	} else {
		fmt.Printf("File: %s\n", fileInfo.Name)
		fmt.Printf("Size: %.0f bytes\n", fileInfo.Size)
		fmt.Printf("Modified: %s\n", fileInfo.ModTime)
	}

	// Optionally analyze the CSV with another Python script
	fmt.Println("\nCreating analysis script...")
	analysisScript := `#!/usr/bin/env python3
import csv
from collections import defaultdict

with open('sales_data.csv', 'r') as f:
    reader = csv.DictReader(f)
    data = list(reader)

# Group by region
region_sales = defaultdict(float)
for row in data:
    region_sales[row['region']] += float(row['total_price'])

print("=== Sales by Region ===")
for region, total in sorted(region_sales.items(), key=lambda x: x[1], reverse=True):
    print(f"{region}: ${total:,.2f}")

# Group by product
product_quantity = defaultdict(int)
for row in data:
    product_quantity[row['product']] += int(row['quantity'])

print("\n=== Top 3 Products by Quantity Sold ===")
for product, qty in sorted(product_quantity.items(), key=lambda x: x[1], reverse=True)[:3]:
    print(f"{product}: {qty} units")
`

	err = client.WriteFile(ctx, sandboxID, "/tmp/analyze_csv.py", []byte(analysisScript))
	if err != nil {
		log.Printf("Failed to write analysis script: %v\n", err)
	} else {
		fmt.Println("✓ Analysis script created")

		// Run analysis
		fmt.Println("\nRunning analysis on generated data...")
		analysisReq := &daytona.ExecuteCommandRequest{
			Command: "python3 /tmp/analyze_csv.py",
			Cwd:     "/tmp",
			Timeout: 10.0,
		}

		analysisResp, err := client.ExecuteCommand(ctx, sandboxID, analysisReq)
		if err != nil {
			log.Printf("Failed to execute analysis: %v\n", err)
		} else {
			fmt.Println(analysisResp.Result)
		}
	}

	// List all files created
	fmt.Println("\nListing all files created in /tmp...")
	files, err := client.ListFiles(ctx, sandboxID, "/tmp")
	if err != nil {
		log.Printf("Failed to list files: %v\n", err)
	} else {
		for _, f := range files {
			if f.Name == "generate_csv.py" || f.Name == "sales_data.csv" || f.Name == "analyze_csv.py" {
				fmt.Printf("  - %s [%.0f bytes]\n", f.Name, f.Size)
			}
		}
	}

	// Clean up - delete the sandbox
	fmt.Println("\nCleaning up...")
	err = client.DeleteSandbox(ctx, sandboxID)
	if err != nil {
		log.Printf("Failed to delete sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox deleted")
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("Successfully generated CSV with mock data, read it, analyzed it, and cleaned up!")
}

#!/usr/bin/env python3
"""Simple data processing script"""

import json
import csv
from datetime import datetime


def load_json(filename):
    """Load data from JSON file"""
    with open(filename, 'r') as f:
        return json.load(f)


def load_csv(filename):
    """Load data from CSV file"""
    data = []
    with open(filename, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            data.append(row)
    return data


def filter_by_category(products, category):
    """Filter products by category"""
    return [p for p in products if p['category'] == category]


def get_in_stock(products):
    """Get products that are in stock"""
    return [p for p in products if p.get('in_stock', '').lower() == 'true']


def calculate_total_value(products):
    """Calculate total inventory value"""
    total = 0.0
    for p in products:
        price = float(p['price'])
        stock = int(p['stock'])
        total += price * stock
    return total


def main():
    print("Data Processing Script")
    print("=" * 30)
    
    # Load products
    products = load_csv('products.csv')
    print(f"Loaded {len(products)} products")
    
    # Show categories
    categories = set(p['category'] for p in products)
    print(f"Categories: {', '.join(categories)}")
    
    # In stock count
    in_stock = get_in_stock(products)
    print(f"Products in stock: {len(in_stock)}")
    
    # Total value
    total_value = calculate_total_value(products)
    print(f"Total inventory value: ${total_value:,.2f}")
    
    # Filter example
    electronics = filter_by_category(products, 'Electronics')
    print(f"Electronics products: {len(electronics)}")


if __name__ == '__main__':
    main()

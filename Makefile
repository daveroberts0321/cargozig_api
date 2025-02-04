# Makefile for CargoZig API

# Variables
PORT ?= 3000

# Run the development server
run:
	@echo "Starting development server on port $(PORT)..."
	@go run main.go

# Git commit and push
git-commit:
	@git add .
	@git commit -m "$(m)"
	@git push

# Usage:
# make develop - to start the server
# make git-commit m="Your commit message" - to commit and push changes 
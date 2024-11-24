# Imolink

Imolink is a real estate platform that integrates with WhatsApp to provide property information and interactions through a chatbot. The platform uses OpenAI for natural language processing and PostgreSQL for data storage and embedding distance calculations.

## Features

- *WhatsApp Integration*: Connect and interact with users via WhatsApp.
- *Property Management*: Add, retrieve, and serve property details.
- *AI-Powered Chatbot*: Use OpenAI to answer user queries and provide property recommendations.
- *Authentication*: Secure API endpoints with token-based authentication.

## Services

### WhatsApp Service

Handles WhatsApp client connections and interactions.

- Connect: /whatsapp/connect - Connects to WhatsApp and provides a QR code for login.
- Reconnect: /whatsapp/reconnect - Reconnects to WhatsApp using stored device information.

### Properties Service

Manages property data and serves property details.

- Add Properties: /properties (POST) - Adds new properties to the database.
- Get Properties: /properties (GET) - Retrieves all properties.
- Serve Property: /properties/:ref - Serves property details as an HTML page.

### Imolink Service

Handles AI interactions and embeddings.

- Ask: /ask (POST) - Processes user questions and provides AI-generated answers.
- Train: /train (POST) - Trains the AI model with new data.
- Purge: /embeddings (DELETE) - Purges all embeddings from the database.

### Auth Service

Provides authentication for API endpoints.

- AuthHandler: Validates bearer tokens for secure access.

### App Service

Utility service for administrative tasks.

- Purge: /purge (POST) - Purges all data from Imolink and Properties services.

## Setup

- Install Docker: Install Docker and Docker Compose on your machine.
- Install Encore: Install the Encore CLI tool for managing services.

## Usage

- Start Services: Run `encore run` to start all services.
- Connect to WhatsApp: Access the `/whatsapp/connect` endpoint to connect the WhatsApp client.
- Manage Properties: Use the `/properties` endpoints to add and retrieve property data.
- Interact with AI: Send messages to the WhatsApp chatbot to ask questions and get suggestions about properties.



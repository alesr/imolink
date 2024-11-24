CREATE TABLE properties (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    area DECIMAL(10,2) NOT NULL,
    num_bedrooms INTEGER NOT NULL,
    num_bathrooms INTEGER NOT NULL,
    num_garage_spots INTEGER NOT NULL,
    price DECIMAL(15,2) NOT NULL,
    
    -- Address fields
    street VARCHAR(255) NOT NULL,
    number INTEGER NOT NULL,
    district VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    state CHAR(2) NOT NULL,
    
    -- Property type
    property_type VARCHAR(50) NOT NULL,
    
    -- Info fields
    reference VARCHAR(50) NOT NULL,
    description TEXT,
    year_built INTEGER NOT NULL,
    builder VARCHAR(255) NOT NULL,
    features VARCHAR[] NOT NULL,
    
    -- Photo fields
    photo_base64_data TEXT,
    photo_format VARCHAR(64),
    photo_upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Blueprint fields
    blueprint_base64_data TEXT,
    blueprint_format VARCHAR(64),
    blueprint_upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS delivery (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) not null,
    phone VARCHAR(20) not null,
    zip VARCHAR(20) not null,
    city VARCHAR(255) not null,
    address VARCHAR(255) not null,
    region VARCHAR(255) not null,
    email VARCHAR(255) not null
);


CREATE TABLE IF NOT EXISTS payment (
    id SERIAL PRIMARY KEY,
    transaction VARCHAR(255) not null,
    request_id VARCHAR(255) not null,
    currency VARCHAR(10) not null,
    provider VARCHAR(255) not null,
    amount INT not null,
    payment_dt BIGINT not null,
    bank VARCHAR(255) not null,
    delivery_cost INT not null,
    goods_total INT not null,
    custom_fee INT not null
);


CREATE TABLE IF NOT EXISTS item (
    id SERIAL PRIMARY KEY,
    chrt_id INT not null,
    track_number VARCHAR(255) not null,
    price INT not null,
    rid VARCHAR(255) not null,
    name VARCHAR(255) not null,
    sale INT not null,
    size VARCHAR(50) not null,
    total_price INT not null,
    nm_id INT not null,
    brand VARCHAR(255) not null,
    status INT not null
);


CREATE TABLE IF NOT EXISTS request_data (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) not null,
    track_number VARCHAR(255) not null,
    entry VARCHAR(255) not null,
    locale VARCHAR(10) not null,
    internal_signature VARCHAR(255) not null,
    customer_id VARCHAR(255) not null,
    delivery_service VARCHAR(255) not null,
    shardkey VARCHAR(255) not null,
    sm_id INT not null,
    date_created TIMESTAMP not null,
    oof_shard VARCHAR(10) not null,
    
    delivery_id INT not null,
    payment_id INT not null,
    item_id INT not null,
    
    FOREIGN KEY (delivery_id) REFERENCES delivery(id),
    FOREIGN KEY (payment_id) REFERENCES payment(id),
    FOREIGN KEY (item_id) REFERENCES item(id),
	full_json text not null
);

select * from request_data

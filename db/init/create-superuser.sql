-- Create superuser if it doesn't exist
CREATE USER IF NOT EXISTS 'superuser'@'%' IDENTIFIED BY 'superpass';
GRANT ALL PRIVILEGES ON *.* TO 'superuser'@'%';
FLUSH PRIVILEGES; 
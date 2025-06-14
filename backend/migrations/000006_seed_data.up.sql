-- Seed sites table
INSERT INTO sites (name, created_at, updated_at) VALUES 
('senti.live', NOW(), NOW()),
('shorts.senti.live', NOW(), NOW()),
('hothinge.com', NOW(), NOW()),
('viblys.com', NOW(), NOW());

-- Seed devices table
INSERT INTO devices (name, created_at, updated_at) VALUES 
('Desktop', NOW(), NOW()),
('Tablet', NOW(), NOW()),
('Mobile', NOW(), NOW());

-- Seed features table
INSERT INTO features (name, created_at, updated_at) VALUES 
('Chat Functionality', NOW(), NOW()),
('Paywall', NOW(), NOW()),
('Age Verification', NOW(), NOW()),
('In-App Bot-First Notifications', NOW(), NOW()),
('Video Playback', NOW(), NOW()),
('In-Video Ads', NOW(), NOW()),
('Dating Profile', NOW(), NOW()),
('Like & Follow', NOW(), NOW()),
('Ad Clicks', NOW(), NOW()),
('Mini Games & Energy Points', NOW(), NOW()),
('iFrame Slot Machine Games', NOW(), NOW()),
('Localisation / Language Support', NOW(), NOW()),
('UI Stability & Consistency', NOW(), NOW()),
('Scrolling Home Page', NOW(), NOW()); 
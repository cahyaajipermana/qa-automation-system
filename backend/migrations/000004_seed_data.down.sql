-- Remove seeded data from features table
DELETE FROM features WHERE name IN (
    'Chat Functionality',
    'Paywall',
    'Age Verification',
    'In-App Bot-First Notifications',
    'Video Playback',
    'In-Video Ads',
    'Dating Profile',
    'Like & Follow',
    'Ad Clicks',
    'Mini Games & Energy Points',
    'iFrame Slot Machine Games',
    'Localisation / Language Support',
    'UI Stability & Consistency'
);

-- Remove seeded data from devices table
DELETE FROM devices WHERE name IN (
    'Desktop',
    'Tablet',
    'Mobile'
);

-- Remove seeded data from sites table
DELETE FROM sites WHERE name IN (
    'senti.live',
    'shorts.senti.live',
    'hothinge.com'
); 
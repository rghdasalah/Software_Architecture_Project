const express = require('express');
     const passport = require('passport');
     const GoogleStrategy = require('passport-google-oauth20').Strategy;
     const jwt = require('jsonwebtoken');
     const redis = require('redis');
     require('dotenv').config();

     const app = express();
     const redisClient = redis.createClient({ url: 'redis://localhost:6379' });
     redisClient.on('error', (err) => console.error('Redis error:', err));
     redisClient.connect();

     passport.use(new GoogleStrategy({
       clientID: process.env.GOOGLE_CLIENT_ID,
       clientSecret: process.env.GOOGLE_CLIENT_SECRET,
       callbackURL: 'http://localhost:8082/auth/google/callback'
     }, (accessToken, refreshToken, profile, done) => {
       const user = { id: profile.id, email: profile.emails[0].value };
       console.log('User profile:', profile);
       done(null, user);
     }));

     app.use(passport.initialize());
     app.get('/auth/google', passport.authenticate('google', { scope: ['profile', 'email'] }));
     app.get('/auth/google/callback', passport.authenticate('google', { session: false }), async (req, res) => {
       const token = jwt.sign({ userId: req.user.id }, process.env.JWT_SECRET, { expiresIn: '1h' });
       await redisClient.setEx(`token:${req.user.id}`, 3600, token);
       console.log(`Stored token for user ${req.user.id}: ${token}`);
       res.json({ token });
     });

     app.listen(8082, () => console.log('Authentication Service running on port 8082'));
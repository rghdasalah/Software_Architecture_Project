
const express = require('express');
     const passport = require('passport');
     const GoogleStrategy = require('passport-google-oauth20').Strategy;
     require('dotenv').config();

     const app = express();

     // Configure Google OAuth2 strategy
     passport.use(new GoogleStrategy({
       clientID: process.env.GOOGLE_CLIENT_ID,
       clientSecret: process.env.GOOGLE_CLIENT_SECRET,
       callbackURL: 'http://localhost:8082/auth/google/callback'
     }, (accessToken, refreshToken, profile, done) => {
       // Simulate user creation/retrieval
       const user = { id: profile.id, email: profile.emails[0].value };
       console.log('User profile:', profile); // Debug: see user data
       done(null, user);
     }));

     // Initialize Passport
     app.use(passport.initialize());

     // Start OAuth2 flow
     app.get('/auth/google', passport.authenticate('google', { scope: ['profile', 'email'] }));

     // Handle callback
     app.get('/auth/google/callback', passport.authenticate('google', { session: false }), (req, res) => {
       res.json({ message: 'Login successful', user: req.user });
     });

     // Start server
     app.listen(8082, () => console.log('Authentication Service running on port 8082'));
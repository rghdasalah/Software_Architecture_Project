# 🚗 Real-Time Chat-Based Ride Coordination Platform

## 📌 Overview
This project is a **real-time chat-based ride coordination platform** designed for students traveling to similar destinations. It enables users to **authenticate securely, communicate in real-time, and track user presence**, ensuring safer and more efficient transportation planning.

## 📜 Features
✅ **User Authentication:** OAuth2-based login with Google.  
✅ **Real-Time Chat:** Instant messaging between users looking for rides.  
✅ **Presence Service:** Tracks online/offline status of users.  
✅ **Microservices Architecture:** Authentication, chat, and presence as independent services.  
✅ **API Gateway:** Secure, unified access to all microservices.  
✅ **Scalable Database:** PostgreSQL with sharding considerations.  
✅ **Caching & Optimization:** Redis for session management and message storage.  
✅ **Monitoring & Logging:** OpenSearch for centralized logging and system observability.  

---

## 🏗️ **System Architecture**
The platform follows a **microservices-based** architecture, consisting of:  

1️⃣ **Authentication Service** – Handles user login via OAuth2 (Google).  
2️⃣ **Real-Time Chat Service** – Manages messaging and chat rooms using WebSockets.  
3️⃣ **Presence Service** – Tracks user availability and online status.  
4️⃣ **API Gateway** – Manages routing, security, and load balancing.  


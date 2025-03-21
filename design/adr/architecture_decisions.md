# ADR 1: Authentication Method – JWT & OAuth2

## Title: Authentication Method for Secure User Access  
**Date:** March 2025  

### Context:
The platform needs a secure authentication mechanism that supports scalability, integration with third-party services, and a seamless user experience.

### Decision:
Use JWT (JSON Web Tokens) for session management and OAuth2 for third-party authentication (Google, Facebook, etc.).

### Alternatives:
- **Session-based authentication** (Less scalable due to session storage issues)
- **API keys** (Less secure for user authentication)

### Consequences:
✅ Stateless authentication for scalability  
✅ Easy third-party authentication integration  
❌ Tokens must be securely stored on the client side  

### Participants:
Backend Developer

### Status:
Accepted  

---

# ADR 2: Database Choice – PostgreSQL & Redis

## Title: Database Selection for Scalability & Performance  
**Date:** March 2025  

### Context:
The system needs structured storage for user data, chat history, and ride coordination, with fast access to frequently used data (e.g., online status, active rides).

### Decision:
- Use **PostgreSQL** for structured user & ride data
- Use **Redis** for real-time presence tracking & caching

### Alternatives:
- **MySQL** (Good alternative but lacks some advanced JSON support)
- **MongoDB** (More suited for unstructured data, but chat history benefits from relational constraints)

### Consequences:
✅ Scalable relational data storage  
✅ Low-latency cache for real-time status updates  
❌ Requires additional setup for Redis  

### Participants:
Database Engineer, Backend Developer

### Status:
Accepted  

---

# ADR 3: Microservices Architecture

## Title: Adopting Microservices for Scalability  
**Date:** March 2025  

### Context:
The platform requires independent scalability for authentication, chat, and presence tracking.

### Decision:
Implement a microservices architecture with separate services for:
- Authentication
- Real-time Chat
- Dashboard

### Alternatives:
- **Monolithic Architecture** (Easier to develop but harder to scale)
- **Serverless Functions** (Scales well but increases cold start latency)

### Consequences:
✅ Independent scaling & deployment  
✅ Better fault isolation  
❌ Increased complexity in service communication  

### Participants:
Software Architect

### Status:
Accepted  

---

# ADR 4: Message Broker – Kafka vs. Redis Pub/Sub

## Title: Choosing a Messaging System for Real-Time Communication  
**Date:** March 2025  

### Context:
The system needs a message broker for real-time chat, presence updates, and notifications.

### Decision:
Use **Redis Pub/Sub** for real-time messaging due to low latency.

### Alternatives:
- **Apache Kafka** (Scalable, durable, but higher latency)
- **RabbitMQ** (More overhead for real-time applications)

### Consequences:
✅ Ultra-fast real-time updates  
✅ Lightweight and easy to deploy  
❌ Not persistent (messages are lost if the system crashes)  

### Participants:
Backend Developer

### Status:
Accepted  


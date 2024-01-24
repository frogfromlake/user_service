# Streamfair User Service

Streamfair (STC) Backend microservice: This service handles operations related to user accounts. It includes creating new accounts, updating account details, and managing user authentication. The service interacts with the Accounts and AccountTypes tables.

#### Backend in progress! 
Currently working on: Providing API

### Backend Tasks

✔️ = partly done / in progress
✅ = done

#### Database Integration:
1. **Implement database support:** PostgreSQL with PGX Driver ✅
2. **Design DB Schema:** dbdiagramm.io ✅
3. **Implement CRUD operations for Tables and Junction tables:** SQL and SQLC -> Go ✅
4. **Implement Unit Tests for CRUD operations:** 80%+ coverage ✅
5. **Implement DB Transactions:** Go ✔️
6. **Implement DB Transaction Unit Tests:** Tabledriven Tests with Go Routines and Channels ✅
7. **Take Care of Transactionlocks & Deadlocks** ✅
8. **Transaction Isolation levels & read phenomena** ✅
9. **Implement Github Actions to run automated Tests:** Go, PostgreSQL ✅
10. **Document DB:** DBDocs


#### Providing API:
1. **Building RESTful HTTP JSON API:** Gin + PASETO ✔️
2. **Loading Configs and Envs:** Go + Viper ✅
3. **Testing HTTP API:** Mock DB / mockgen (source) ✔️
4. **DB Error handling:** tags and custom validators ✔️
5. **Strengthen Unit Tests:** Custom gomock matcher
8. **Implement Login User API:** PASETO + Auth Middleware (Gin)
6. **Middleware:** Authentication middleware, authorization rules gin
7. **Storing sensitive Data:** Hashing with Bcrypt
8. **Refactoring:** Unifying codebase, improve security of codebase and
   improve performance of db interactions and api ✔️

#### Deployment:
1. **Deploying to production:** Docker + Kubernetes + AWS(?)
2. **Some more Cloud Engineering stuff**


#### Advanced Backend Stuff:
1. **Implement User sessions:** refresh tokens
2. **Implement gRPC API:** protobuf
3. **Implement gRPC Gateway for serving gRPC & HTTP requests**
4. **Document API:** ReDocly, Postman or SwaggerUI from go server
5. **Some more advanced backend stuff**
6. **Implement authorization to protect gRPC API**
7. **Improve logging for gRPC API**
8. **Write HTTP logger middleware:** Go


#### Asynchronous processing:
1. **Async processing with background workers:** Asynq + Redis
2. **Email verification API**
3. **Implement Unit Tests for gRPC API:** mockdb + redis
4. **Unit testing with authentication**


#### Improving Stability and Security of the Server:
1. **To be updated**

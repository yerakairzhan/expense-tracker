# Practice 7 — Authentication & Authorization in Go

## Before starting
- x/crypto/bcrypt (official docs)
- golang-jwt (official docs)
- Authentication & Authorization with Go using JWT

---

## Project Structure
```
cmd/
config/
internal/
  app/
  controller/http/v1/
    error.go
    router.go
    user.go
  entity/
    DTO.go
    user.go
  usecase/
    repo/
    interfaces.go
    user.go
pkg/
utils/
.env
```

---

## Step 1: Define User struct
```go
type User struct {
    ID uuid.UUID
    Username string
    Email string
    Password string
    Role string
    Verified bool
}
```

---

## Step 2: Register User

### Endpoint
POST /users

### DTO
```go
type CreateUserDTO struct {
    Username string `json:"username" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Role string `json:"role"`
}
```

### Handler
```go
func (r *userRoutes) RegisterUser(c *gin.Context) {
    var dto entity.CreateUserDTO

    if err := c.ShouldBindJSON(&dto); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    hashedPassword, _ := utils.HashPassword(dto.Password)

    user := entity.User{
        Username: dto.Username,
        Email: dto.Email,
        Password: hashedPassword,
        Role: "user",
    }

    createdUser, sessionID, _ := r.t.RegisterUser(&user)

    c.JSON(http.StatusCreated, gin.H{
        "message": "User registered successfully",
        "session_id": sessionID,
        "user": createdUser,
    })
}
```

---

## Utils
```go
func HashPassword(password string) (string, error) {
    return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}
```

---

## Step 3: Use Case
```go
type UserInterface interface {
    RegisterUser(user *entity.User) (*entity.User, string, error)
}
```

```go
func (u *UserUseCase) RegisterUser(user *entity.User) (*entity.User, string, error) {
    user, err := u.repo.RegisterUser(user)
    if err != nil {
        return nil, "", err
    }
    sessionID := uuid.New().String()
    return user, sessionID, nil
}
```

---

## Step 4: Login

### Endpoint
POST /users/login

### DTO
```go
type LoginUserDTO struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}
```

### Handler
```go
func (r *userRoutes) LoginUser(c *gin.Context) {
    var input entity.LoginUserDTO
    c.ShouldBindJSON(&input)

    token, _ := r.t.LoginUser(&input)
    c.JSON(200, gin.H{"token": token})
}
```

---

## Login Logic
```go
func (u *UserUseCase) LoginUser(user *entity.LoginUserDTO) (string, error) {
    userFromRepo, _ := u.repo.LoginUser(user)

    if !utils.CheckPassword(userFromRepo.Password, user.Password) {
        return "", nil
    }

    return utils.GenerateJWT(userFromRepo.ID, userFromRepo.Role)
}
```

---

## JWT + Password
```go
func CheckPassword(hash, password string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateJWT(userID uuid.UUID, role string) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "role": role,
        "exp": time.Now().Add(24 * time.Hour).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
```

---

## Step 5: Middleware
```go
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenStr := c.GetHeader("Authorization")

        if tokenStr == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Token required"})
            return
        }

        tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }

        claims := token.Claims.(jwt.MapClaims)
        c.Set("userID", claims["user_id"])
        c.Set("role", claims["role"])

        c.Next()
    }
}
```

---

## Protected Route
```go
protected := h.Group("/")
protected.Use(utils.JWTAuthMiddleware())

protected.GET("/protected/hello", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "OK"})
})
```

---

## Assignment

### 1. GetMe
- /users/me
- Use JWT only

### 2. Role-Based Access
- PATCH /users/promote/:id
- Only admin

### 3. Rate Limiter
- Limit requests
- JWT → userID
- No JWT → IP

---

## Deliverables
- GitHub repo
- Demo video

---

## Deadline
12.04.2026 23:59

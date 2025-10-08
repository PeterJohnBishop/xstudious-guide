# xstudious-guide

Auth (JWT) → secure, user-aware endpoints
File storage (S3) → images, documents, media uploads
Data store (DynamoDB) → scalable, flexible schema storage
Realtime comms (WebSocket Hub) → chat, notifications, live dashboards
Email delivery (Resend) → transactional or marketing emails
3rd-party integrations (webhooks, Google Maps) → extendable to external services

POST {{baseUrl}}/register

Request:
{
  "name": "Peter Bishop",
  "email": "pjb.den@gmail.com",
  "password": ************
}
Response:
{
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IiIsIm5hbWUiOiJQZXRlciBCaXNob3AiLCJlbWFpbCI6InBqYi5kZW5AZ21haWwuY29tIiwidG9rZW5fdHlwZSI6ImFjY2VzcyIsImV4cCI6MTc1OTg2NDMzMSwiaWF0IjoxNzU5ODYzNDMxfQ.tFyCKZBAoPPf10R_9M1vuStVgsuOVAsvfNWqIO4ZljI",
    "message": "User created successfully",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjA0NjgyMzEsImlhdCI6MTc1OTg2MzQzMX0.ZN30REvje7_f98Qhg-4uHxIiV0ZfsqHTzFGQQlEjGyU",
    "user.id": "u_8uqeJRURJC0ZoYYpqlJw"
}

POST {{baseUrl}}/login

Request:
{
  "email": "pjb.den@gmail.com",
  "password": ************
}
Response:
{
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InVfOHVxZUpSVVJKQzBab1lZcHFsSnciLCJuYW1lIjoiUGV0ZXIgQmlzaG9wIiwiZW1haWwiOiJwamIuZGVuQGdtYWlsLmNvbSIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJleHAiOjE3NTk4NjQ0MDMsImlhdCI6MTc1OTg2MzUwMywic3ViIjoidV84dXFlSlJVUkpDMFpvWVlwcWxKdyJ9.-GUcKNowzGmy4rFiXUmHTuwv8SlQGyIHNxPLQNUUM-g",
    "message": "Login Success",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjA0NjgzMDMsImlhdCI6MTc1OTg2MzUwMywic3ViIjoidV84dXFlSlJVUkpDMFpvWVlwcWxKdyJ9.n5K4IletNzli5cVcNO5AnDyVzH1OiFE0uX823Y-FWd0",
    "user": {
        "id": "u_8uqeJRURJC0ZoYYpqlJw",
        "name": "Peter Bishop",
        "email": "pjb.den@gmail.com",
        "password": "$2a$10$lVHu2d3aHeo3ht1h8Go0auQePwARCXKDYSvwJNAvtNjfOxxSvnYle"
    }
}

POST {{baseUrl}}/refresh-token

Request:
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjA0Njg0NDEsImlhdCI6MTc1OTg2MzY0MSwic3ViIjoidV84dXFlSlJVUkpDMFpvWVlwcWxKdyJ9.jX4E9J3XCyxzV0azwS37xwSrt3EQxFuJvAocdl6ZoCo"
}
Response:
{
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InVfOHVxZUpSVVJKQzBab1lZcHFsSnciLCJuYW1lIjoiUGV0ZXIgQmlzaG9wIiwiZW1haWwiOiJwamIuZGVuQGdtYWlsLmNvbSIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJleHAiOjE3NTk4NjQ2OTksImlhdCI6MTc1OTg2Mzc5OSwic3ViIjoidV84dXFlSlJVUkpDMFpvWVlwcWxKdyJ9.j7Dzqji0Fi5uyMYkR7BEQs3LG2fpZYdM6xermJgaU40"
}

GET {{baseUrl}}/users

Response: 
{
    "message": "Users Found!",
    "users": [
        {
            "id": "u_iyFxpKkgQkCKztcAr9183Q",
            "name": "UpdatedUser",
            "email": "test1@gmail.com",
            "password": "$2a$10$UUtsFr6VEoShzwPqUeaRb.syO4kCX/xGzkN8kCjPqpbOrsND4iJV."
        },
        {
            "id": "u_dz6zSziQQGGBMV7diDoqlQ",
            "name": "test2",
            "email": "test2@gmail.com",
            "password": "$2a$10$NT6/zkrr2qX5tgWcdWJqCOpnhB10zFwbkuyqZ5ech5eyb.Pgb3zhG"
        },
        {
            "id": "u_8uqeJRURJC0ZoYYpqlJw",
            "name": "Peter Bishop",
            "email": "pjb.den@gmail.com",
            "password": "$2a$10$lVHu2d3aHeo3ht1h8Go0auQePwARCXKDYSvwJNAvtNjfOxxSvnYle"
        }
    ]
}

GET {{baseUrl}}/users/{{user.id}}

Response:
{
    "message": "User Found!",
    "user": {
        "id": "u_dz6zSziQQGGBMV7diDoqlQ",
        "name": "test2",
        "email": "test2@gmail.com",
        "password": "$2a$10$NT6/zkrr2qX5tgWcdWJqCOpnhB10zFwbkuyqZ5ech5eyb.Pgb3zhG"
    }
}

PUT {{baseUrl}}/users

Request:
{
  "id": "u_dz6zSziQQGGBMV7diDoqlQ",
  "name": "Test 2 Updated",
  "email": "pbsihop+2@clickup.com"
}
Response:
{
    "message": "User Updated!"
}

DELETE {{baseUrl}}/users/{{user.id}}

Response: 
{
    "message": "User Deleted!"
}
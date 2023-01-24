## Simple Go Web API

##### SHORT DESCRIPTION
This simple web api is written in Go. It exposes four api endpoints with paths /add, /subtract, 
/multiply and /divide. Each of those endpoints takes input x and y as query parameters and returns
a result of mathematical operation on those operands. Results are cached for one minute.
###### Note: x and y are restricted to integer values.

##### STEPS TO RUN ON MAC AND LINUX
* open terminal window
* ```cd``` into project's root directory
* ```go run main.go``` will run API locally on port 8080
* type ```curl "localhost:8080/multiply?x=2&y=3"``` in terminal window
* expected output: {"action":"multiply","x":2,"y":3,"answer":"6","cacheUsed":false} 
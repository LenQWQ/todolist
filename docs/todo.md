### 1. N/A

1. route definition

- Url: /todos
- Method: POST
- Request: `CreateTodoReq`
- Response: `Todo`

2. request definition



```golang
type CreateTodoReq struct {
	Titles string `json:"titles"`
	Content string `json:"content"`
}
```


3. response definition



```golang
type Todo struct {
	Id int `json:"id"`
	Titles string `json:"titles"`
	Done bool `json:"done"`
	Content string `json:"content"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}
```

### 2. N/A

1. route definition

- Url: /todos
- Method: GET
- Request: `-`
- Response: `TodoListResp`

2. request definition



3. response definition



```golang
type TodoListResp struct {
	Todos []Todo `json:"todos"`
}
```

### 3. N/A

1. route definition

- Url: /todos/:id
- Method: GET
- Request: `GetTodoReq`
- Response: `Todo`

2. request definition



```golang
type GetTodoReq struct {
	Id int `path:"id"`
}
```


3. response definition



```golang
type Todo struct {
	Id int `json:"id"`
	Titles string `json:"titles"`
	Done bool `json:"done"`
	Content string `json:"content"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}
```

### 4. N/A

1. route definition

- Url: /todos/:id
- Method: PUT
- Request: `UpdateTodoReq`
- Response: `Todo`

2. request definition



```golang
type UpdateTodoReq struct {
	Id int `path:"id"`
	Titles string `json:"titles"`
	Content string `json:"content"`
	Done bool `json:"done"`
}
```


3. response definition



```golang
type Todo struct {
	Id int `json:"id"`
	Titles string `json:"titles"`
	Done bool `json:"done"`
	Content string `json:"content"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}
```

### 5. N/A

1. route definition

- Url: /todos/:id
- Method: DELETE
- Request: `DeleteTodoReq`
- Response: `-`

2. request definition



```golang
type DeleteTodoReq struct {
	Id int `path:"id"`
}
```


3. response definition



package graphql

import (
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

// Handler returns an HTTP handler for the GraphQL API
func Handler(resolver *Resolver) http.Handler {
	schema := graphql.MustParseSchema(Schema, resolver)
	return &relay.Handler{Schema: schema}
}

// PlaygroundHandler returns an HTTP handler for the GraphQL playground UI
func PlaygroundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Simple HTML page with GraphQL playground
		playground := `
<!DOCTYPE html>
<html>
<head>
  <meta charset=utf-8/>
  <meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
  <title>GAT API GraphQL Playground</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/static/css/index.css" />
  <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/favicon.png" />
  <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/static/js/middleware.js"></script>
</head>
<body>
  <div id="root">
    <style>
      body {
        background-color: rgb(23, 42, 58);
        font-family: Open Sans, sans-serif;
        height: 90vh;
      }
      #root {
        height: 100%;
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
      }
      .loading {
        font-size: 32px;
        font-weight: 200;
        color: rgba(255, 255, 255, .6);
        margin-left: 20px;
      }
      img {
        width: 78px;
        height: 78px;
      }
      .title {
        font-weight: 400;
      }
    </style>
    <img src='https://cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/logo.png' alt=''>
    <div class="loading"> Loading
      <span class="title">GAT API GraphQL Playground</span>
    </div>
  </div>
  <script>window.addEventListener('load', function (event) {
      GraphQLPlayground.init(document.getElementById('root'), {
        endpoint: '/graphql',
        settings: {
          'request.credentials': 'same-origin',
        }
      })
    })</script>
</body>
</html>
`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(playground))
	})
}

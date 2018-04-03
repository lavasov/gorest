Experiments with Go and some popular Go libraries.

Scope:  

- [x] Use echo or any other tool a web handler  
- [ ] Implement login endpoint with JWT token and simple middleware that checks header for 'Authorization: Bearer %jwt_token%' in each request. Otherwise return 403 and json struct with error.
- [ ] Implement endpoint that will use OAuth2 authorization for FB to login and issue access_token
- [x] Log each request including status code using logrus or https://github.com/uber-go/zap
- [x] Implement persistence with Gorm
- [x] Use tool of your choice for DB migrations
- [x] Implement save endpoint for Task object
- [x] Implement update endpoint for Task object
- [x] Implement get endpoint for Task object
- [x] Implement delete endpoint for Task object (just update IsDeleted field)
- [x] Use CORS (reply with header Access-Control-Allow-Origin: *)
- [x] Add support for OPTION HTTP method for each endpoints
- [x] Configure daemon over simple YAML config. Specify path as process flag for daemon. Required params: ListenAddress, DatabaseUri.
- [x] Use vendoring
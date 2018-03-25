Experiments with Go and some popular Go libraries.

Scope:  

- [x] Use echo or any other tool a web handler  
- [ ] Implement login endpoint with JWT token and simple middleware that checks header for 'Authorization: Bearer %jwt_token%' in each request. Otherwise return 403 and json struct with error.
- [ ] Implement endpoint that will use OAuth2 authorization for FB to login and issue access_token
- [ ] Log each request including status code (logrus or https://github.com/uber-go/zap or someyhing else with good rating on github)
- [x]Implement persistence with Gorm (https://github.com/jinzhu/gorm or sqlx)
- [ ] Use tool of your choice for DB migrations
- [ ] Implement save endpoint for Task object
- [ ] Implement update endpoint for Task object
- [ ] Implement get endpoint for Task object
- [ ] Implement delete endpoint for Task object (just update IsDeleted field)
- [ ] Use CORS (reply with header Access-Control-Allow-Origin: *)
- [ ] Add support for OPTION HTTP method for each endpoints
- [x] Configure daemon over simple YAML config. Specify path as process flag for daemon. Required params: ListenAddress, DatabaseUri.
- [x] Use vendoring
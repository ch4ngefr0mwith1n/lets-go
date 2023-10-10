package main

type contextKey string

// u ovoj promjenjivoj će se čuvati ključ koji je vezan za "authentication status"
// njega koristimo za čuvanje i vađenje "authentication status"-a
const isAuthenticatedContextKey = contextKey("isAuthenticated")

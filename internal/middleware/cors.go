package middleware

//func CORS(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
//		w.Header().Set("Access-Control-Allow-Credentials", "true")
//
//		if r.Method == http.MethodOptions {
//			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, User-Agent, X-CSRF-TOKEN")
//			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
//			w.Header().Set("Access-Control-Max-Age", "86400")
//			w.WriteHeader(http.StatusOK)
//		} else {
//			next.ServeHTTP(w, r)
//		}
//	})
//}

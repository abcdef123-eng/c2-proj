package server

import (
        "time"
        "fmt"
        "github.com/golang-jwt/jwt/v5"
        "errors"
)


func (s *Server) ValidateToken(token string) (bool, error, uint32) {
        JwtTokenSecret := s.Config.JWTSecret

        tok, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
                        // Ensure it's using the expected signing method
                        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                                return false, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
                        }
                        return []byte(JwtTokenSecret), nil
                })
        if err != nil {
            switch {
            case errors.Is(err, jwt.ErrTokenExpired):
                return false, err, ERROR_TOKEN_EXPIRED
            case errors.Is(err, jwt.ErrTokenSignatureInvalid):
                return false, err, ERROR_TOKEN_SIGNWRONG
            case errors.Is(err, jwt.ErrTokenMalformed):
                return false, err, ERROR_TOKEN_INVALID
            default:
                return false, err, ERROR_TOKEN_INVALID
            }
        }

        claims, ok := tok.Claims.(jwt.MapClaims)
        if !ok || !tok.Valid {
                return false, fmt.Errorf("Failed Getting Claims"), ERROR_INVALID_CLAIMS

        }      

        user, ok1 := claims["username"].(string)
        host, ok2 := claims["hostname"].(string)
        
        if !ok1 || !ok2 {
                return false,  fmt.Errorf("Invalid Calims", err), ERROR_INVALID_CLAIMS
        }

        ok, err = s.LookupIfUserExists(user, host)
        if err != nil {
                return false, err, ERROR_SERVER_ERROR
        } else if !ok {
                return false, fmt.Errorf("User Doesnt Exsit", err), ERROR_INVALID_CLAIMS
        }

        return true, nil, 0




        


}

func (s *Server) CreateJWT(username string, hostname string) (string, string, error) {
        JwtTokenSecret := s.Config.JWTSecret
        JwtRefreshSecret := s.Config.JWTRefreshSecret

        claims := jwt.MapClaims {
                "username": username,
                "hostname": hostname,
                "ext": time.Now().Add(14 * time.Hour).Unix(),
        }

        at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        accessToken, err := at.SignedString([]byte(JwtTokenSecret))
        if err != nil {
                return "", "", err
        }
        claims2 := jwt.MapClaims {
                "username": username,
                "hostname": hostname,
                "ext": time.Now().Add(14 * 24 * time.Hour).Unix(),
        }

        rt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims2)
        refrehsToken, err := rt.SignedString([]byte(JwtRefreshSecret))
        if err != nil {
                return "", "", err
        }

        return accessToken, refrehsToken, nil


}

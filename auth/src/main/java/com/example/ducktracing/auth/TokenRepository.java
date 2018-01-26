package com.example.ducktracing.auth;

/**
 * @author Beka Tsotsoria
 */
public interface TokenRepository {

    Token findByValue(String value);

    void add(Token token);
}

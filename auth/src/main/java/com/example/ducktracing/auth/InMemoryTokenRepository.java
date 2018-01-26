package com.example.ducktracing.auth;

import org.springframework.stereotype.Repository;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * @author Beka Tsotsoria
 */
@Repository
public class InMemoryTokenRepository implements TokenRepository {

    private final Map<String, Token> tokens = new ConcurrentHashMap<>();

    @Override
    public Token findByValue(String value) {
        try {
            Thread.sleep(1000);
        } catch (InterruptedException ignored) {
        }
        return tokens.get(value);
    }

    @Override
    public void add(Token token) {
        tokens.put(token.getValue(), token);
    }
}

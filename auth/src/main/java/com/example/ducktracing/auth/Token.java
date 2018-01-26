package com.example.ducktracing.auth;

import java.time.LocalDateTime;
import java.time.temporal.ChronoUnit;
import java.util.Base64;

/**
 * @author Beka Tsotsoria
 */
public class Token {

    private String owner;
    private String value;
    private LocalDateTime createdAt;

    public Token(String owner, String value, LocalDateTime createdAt) {
        this.owner = owner;
        this.value = value;
        this.createdAt = createdAt;
    }

    public String getOwner() {
        return owner;
    }

    public String getValue() {
        return value;
    }

    public LocalDateTime getCreatedAt() {
        return createdAt;
    }

    public boolean isValid() {
        return createdAt.until(LocalDateTime.now(), ChronoUnit.MINUTES) < 10;
    }

    public String getBasicAuthHeaderValue() {
        return "Basic " + Base64.getEncoder().encodeToString((owner + ":" + value).getBytes());
    }

    @Override
    public String toString() {
        return "Token{" +
            "value='" + value + '\'' +
            ", createdAt=" + createdAt +
            '}';
    }
}

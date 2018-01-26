package com.example.ducktracing.auth;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.support.RedirectAttributes;

import java.time.LocalDateTime;
import java.util.UUID;

/**
 * @author Beka Tsotsoria
 */
@Controller
@RequestMapping("/tokens")
public class TokenController {

    private TokenRepository tokenRepository;

    @Autowired
    public TokenController(TokenRepository tokenRepository) {
        this.tokenRepository = tokenRepository;
    }

    @RequestMapping(method = RequestMethod.GET)
    public String index() {
        return "index";
    }

    @RequestMapping(value = "/{value}", method = RequestMethod.GET)
    @ResponseBody
    public Token getToken(@PathVariable String value) {
        return tokenRepository.findByValue(value);
    }

    @RequestMapping(method = RequestMethod.POST)
    public String generateToken(@RequestParam String owner,
                                @RequestParam(required = false) String redirectUrl,
                                @RequestParam(required = false) String authHeaderName,
                                RedirectAttributes attrs) {
        Token token = new Token(owner, UUID.randomUUID().toString(), LocalDateTime.now());
        tokenRepository.add(token);
        attrs.addFlashAttribute("token", token);
        attrs.addFlashAttribute("redirectUrl", redirectUrl);
        attrs.addFlashAttribute("authHeaderName", authHeaderName);
        return "redirect:/tokens";
    }

}

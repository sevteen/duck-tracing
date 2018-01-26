package com.example.ducktracing.auth;

import io.opentracing.Span;
import io.opentracing.SpanContext;
import io.opentracing.Tracer;
import io.opentracing.propagation.Format;
import io.opentracing.propagation.TextMapExtractAdapter;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.support.RedirectAttributes;

import javax.servlet.http.HttpServletRequest;
import java.time.LocalDateTime;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

/**
 * @author Beka Tsotsoria
 */
@Controller
@RequestMapping("/tokens")
public class TokenController {

    private TokenRepository tokenRepository;
    private HttpServletRequest httpRequest;
    private Tracer tracer;

    @Autowired
    public TokenController(TokenRepository tokenRepository, HttpServletRequest httpRequest, Tracer tracer) {
        this.tokenRepository = tokenRepository;
        this.httpRequest = httpRequest;
        this.tracer = tracer;
    }

    @RequestMapping(method = RequestMethod.GET)
    public String index() {
        return "index";
    }

    @RequestMapping(value = "/{value}", method = RequestMethod.GET)
    @ResponseBody
    public Token getToken(@PathVariable String value) {
        Span span = newSpanBuilder("getToken")
            .withTag("tokenValue", value)
            .start();

        Token token = tokenRepository.findByValue(value);

        span.finish();
        return token;
    }

    @RequestMapping(method = RequestMethod.POST)
    public String generateToken(@RequestParam String owner,
                                @RequestParam(required = false) String redirectUrl,
                                @RequestParam(required = false) String authHeaderName,
                                RedirectAttributes attrs) {
        Span span = newSpanBuilder("generateToken")
            .withTag("owner", owner)
            .start();

        Token token = new Token(owner, UUID.randomUUID().toString(), LocalDateTime.now());
        tokenRepository.add(token);
        attrs.addFlashAttribute("token", token);
        attrs.addFlashAttribute("redirectUrl", redirectUrl);
        attrs.addFlashAttribute("authHeaderName", authHeaderName);

        span.finish();
        return "redirect:/tokens";
    }

    private Tracer.SpanBuilder newSpanBuilder(String operationName) {
        Map<String, String> headers = new HashMap<>();
        Collections.list(httpRequest.getHeaderNames())
            .forEach(h -> headers.put(h, httpRequest.getHeader(h)));
        SpanContext context = tracer.extract(Format.Builtin.HTTP_HEADERS, new TextMapExtractAdapter(headers));
        return tracer.buildSpan(operationName)
            .asChildOf(context);
    }

}

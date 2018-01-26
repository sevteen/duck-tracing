package com.example.ducktracing.auth;

import com.uber.jaeger.Configuration;
import com.uber.jaeger.samplers.ConstSampler;
import io.opentracing.Tracer;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;


/**
 * @author Beka Tsotsoria
 */
@org.springframework.context.annotation.Configuration
public class JaegerConfig {

    @Value("${jaeger.service-name}")
    private String serviceName;

    @Value("${jaeger.host-port\\:localhost:6831}")
    private String hostPort;

    @Bean(destroyMethod = "close")
    public Tracer tracer() {
        Configuration config = new Configuration(serviceName,
            new Configuration.SamplerConfiguration(ConstSampler.TYPE, 1),
            new Configuration.ReporterConfiguration(true,
                getAgentHost(),
                getAgentPort(),
                null, null));
        return config.getTracer();
    }

    private String getAgentHost() {
        return hostPort.split(":")[0];
    }

    private int getAgentPort() {
        return Integer.parseInt(hostPort.split(":")[1]);
    }
}

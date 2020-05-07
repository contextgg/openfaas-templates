const express = require('express');
const { createLightship } = require('lightship');
const { register } = require('prom-client');

const handler = require('./function/handler.js');

const PORT = process.env.PORT || 8080;
const METRICS_PORT = process.env.METRICS_PORT || 8081;
const HEALTH_PORT = process.env.HEALTH_PORT || 8082;

const runServer = async () => {
  const app = express();
  const metrics = express();
  
  const lightship = createLightship({
    detectKubernetes: false,
    port: HEALTH_PORT,
  });
  lightship.registerShutdownHandler(() => {
    app.close();
  });
  
  metrics.get("/metrics", ((_, res) => res.send(register.metrics())));

  await handler(app);

  app.listen({ port: PORT }, (() => {
  }));
  metrics.listen({ port: METRICS_PORT }, (() => {
  }));
  
  // when we are ready!
  lightship.signalReady();
};

try {
  runServer();
} catch (err) {
  console.error(err);
}
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
  
  metrics.get("/metrics", ((_, res) => res.send(register.metrics())));

  await handler(app);

  const appServer = app.listen(PORT);
  const metricsServer = metrics.listen(METRICS_PORT);
  
  const lightship = await createLightship({
    port: HEALTH_PORT,
  });
  lightship.registerShutdownHandler(() => {
    appServer.close();
    metricsServer.close();
  });

  // when we are ready!
  lightship.signalReady();
};

try {
  runServer();
} catch (err) {
  console.error(err);
}
const next = require('next')

const dev = process.env.NODE_ENV !== 'production'
const app = next({ dev, dir: __dirname })
const handle = app.getRequestHandler()

const handler = async (server) => {
  await app.prepare();

  server.all('*', (req, res) => {
    return handle(req, res)
  })
};

module.exports = handler
const { HttpLink } = require('apollo-link-http');
const { ApolloServer } = require('apollo-server-express');
const {
  makeRemoteExecutableSchema,
  introspectSchema,
  mergeSchemas
} = require('graphql-tools');
const fetch = require('node-fetch');

// graphql API metadata
const APIS = process.env.APIS || 'https://api.context.gg/graphql';

// create executable schemas from remote GraphQL APIs
const createRemoteExecutableSchemas = async () => {
  const apis = APIS.split(',');

  let schemas = [];
  for (const api of apis) {
    const link = new HttpLink({
      uri: api,
      fetch
    });
    const remoteSchema = await introspectSchema(link);
    const remoteExecutableSchema = makeRemoteExecutableSchema({
      schema: remoteSchema,
      link
    });
    schemas.push(remoteExecutableSchema);
  }
  return schemas;
};

const createNewSchema = async () => {
  const schemas = await createRemoteExecutableSchemas();
  return mergeSchemas({
    schemas
  });
};

const handler = async (cfg) => {
  // Get newly merged schema
  const schema = await createNewSchema();

  // start server with the new schema
  return new ApolloServer({
    schema,
    engine: false,
    ...cfg,
  });
};

module.exports = handler
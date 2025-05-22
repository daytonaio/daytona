module.exports = {
    proto: './apps/runner-grpc/proto/runner.proto',
    outDir: './libs/runner-grpc-client/src',
    generateClient: true,
    generateServer: false,
    clientOptions: {
      includeDocs: true,
      includeTypeScript: true,
      promisifyMethods: true
    }
  };
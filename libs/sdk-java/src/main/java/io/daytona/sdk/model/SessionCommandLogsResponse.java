package io.daytona.sdk.model;

public class SessionCommandLogsResponse extends io.daytona.toolbox.client.model.SessionCommandLogsResponse {
    public static SessionCommandLogsResponse from(io.daytona.toolbox.client.model.SessionCommandLogsResponse resp) {
        SessionCommandLogsResponse result = new SessionCommandLogsResponse();
        result.setOutput(resp.getOutput());
        result.setStdout(resp.getStdout());
        result.setStderr(resp.getStderr());
        return result;
    }
}

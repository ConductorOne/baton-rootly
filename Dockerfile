FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-rootly"]
COPY baton-rootly /
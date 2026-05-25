FROM nousresearch/hermes-agent:main

# Install hermes-claude-auth (https://github.com/kristianvast/hermes-claude-auth)
# The sitecustomize MetaPathFinder hook doesn't fire reliably in this environment,
# so we append bypass activation directly to anthropic_adapter.py — it runs at
# module import time without depending on the hook mechanism.
RUN git clone --depth=1 https://github.com/kristianvast/hermes-claude-auth.git /tmp/hermes-claude-auth && \
    mkdir -p /opt/hermes-claude-auth && \
    cp /tmp/hermes-claude-auth/anthropic_billing_bypass.py /opt/hermes-claude-auth/ && \
    chown -R hermes:hermes /opt/hermes-claude-auth && \
    chmod 644 /opt/hermes-claude-auth/anthropic_billing_bypass.py && \
    rm -rf /tmp/hermes-claude-auth

ENV HERMES_PATCHES_DIR=/opt/hermes-claude-auth

# Append bypass activation to anthropic_adapter.py. At the end of the file,
# sys.modules[__name__] already holds the (partially-populated) module object
# because Python registers it before executing the module body.
RUN printf '\n# hermes-claude-auth: activate billing bypass at module import time\nimport sys as _hca_sys\n_hca_sys.path.insert(0, "/opt/hermes-claude-auth")\ntry:\n    import anthropic_billing_bypass as _hca_bypass\n    _hca_bypass.apply_patches(_hca_sys.modules.get(__name__))\nexcept Exception as _hca_e:\n    _hca_sys.stderr.write("[hermes-claude-auth] bypass failed: " + str(_hca_e) + "\\n")\n' >> /opt/hermes/agent/anthropic_adapter.py

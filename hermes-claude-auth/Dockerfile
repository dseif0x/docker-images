FROM nousresearch/hermes-agent:main

# Install hermes-claude-auth (https://github.com/kristianvast/hermes-claude-auth)
# Patch and hook go into non-volume paths so they survive /opt/data volume mounts.
# HERMES_PATCHES_DIR tells the sitecustomize hook where to find the patch.
RUN git clone --depth=1 https://github.com/kristianvast/hermes-claude-auth.git /tmp/hermes-claude-auth && \
    mkdir -p /opt/hermes-claude-auth && \
    cp /tmp/hermes-claude-auth/anthropic_billing_bypass.py /opt/hermes-claude-auth/ && \
    SITE_PACKAGES=$(/opt/hermes/.venv/bin/python -c "import site; print(site.getsitepackages()[0])") && \
    cp /tmp/hermes-claude-auth/sitecustomize_hook.py "$SITE_PACKAGES/sitecustomize.py" && \
    chmod 644 "$SITE_PACKAGES/sitecustomize.py" /opt/hermes-claude-auth/anthropic_billing_bypass.py && \
    chown -R hermes:hermes /opt/hermes-claude-auth && \
    rm -rf /tmp/hermes-claude-auth

ENV HERMES_PATCHES_DIR=/opt/hermes-claude-auth

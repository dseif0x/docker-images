FROM continuumio/miniconda3:latest
ENV DEBIAN_FRONTEND=noninteractive

ENV NB_USER=jovyan \
    NB_UID=1000 \
    HOME=/home/jovyan
RUN adduser \
        --disabled-password \
        --gecos "Default user" \
        --uid ${NB_UID} \
        --home ${HOME} \
        --force-badname \
        ${NB_USER}

RUN apt-get update \
 && apt-get upgrade --yes \
 && apt-get install --yes --no-install-recommends \
        ca-certificates \
        dnsutils \
        iputils-ping \
        tini \
        # requirement for nbgitpuller
        git \
 && rm -rf /var/lib/apt/lists/*

 # Install dependencies
RUN conda install -n base -c conda-forge nb_conda_kernels \
        jupyterhub \
        jupyterlab \
        nbclassic \
        nbgitpuller \
        ipykernel \
        notebook \
&& conda clean -afy

WORKDIR ${HOME}
USER ${NB_USER}

COPY .condarc .

EXPOSE 8888
ENTRYPOINT ["tini", "--"]
CMD ["jupyter", "lab"]
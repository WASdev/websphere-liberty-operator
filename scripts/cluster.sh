#!/bin/bash

usage() {
  cat <<EOF
$0: Configure a Fyre OCP cluster for end-to-end testing of operators. You can specify the cluster via parameters or just login to it via 'oc login' first.

Examples:

  Configure a cluster:
    $0 <cluster> -u <username> -p <password> -k <apikey> -A
    $0 api.example.cp.fyre.ibm.com:6443 -u kubeadmin -p 1234-5678-91011 -k abcdefghijklmnop -A
    $0 api.example.cp.fyre.ibm.com:6443 -u kubeadmin -p 1234-5678-91011 -k abcdefghijklmnop -A -U adminuser:myp@sswd

  Perform a single configuration step:
    $0 <cluster> -u <username> -p <password> -<step-option>
    $0 api.example.cp.fyre.ibm.com:6443 -u kubeadmin -p 1234-5678-91011 -R
    $0 api.example.cp.fyre.ibm.com:6443 -u kubeadmin -p 1234-5678-91011 -I

  Configuring a cluster you're already logged into (via oc login):
    oc login api.example.cp.fyre.ibm.com:6443 -u kubeadmin -p 1234-5678-91011
    $0 -k abcdefghijklmnop -A

Arguments:
  -u|--user - The username to log in to the OCP cluster. Defaults to 'kubeadmin'.
  -p|--pass - The password to log in to the OCP cluster. Required unless already logged in.
  -k|--key  - The API key for IBM's Staging Container Registry (cp.stg.icr.io). 
                Required if -P or -A if used (to add the pull secret to the cluster's .dockerconfigjson.)
  -A|--all  - Completely configure the cluster by performing all configuration steps (except for 
                --add-cluster-admin). Equivalent to '-R -S -K -L -I -P'.
  -h|--help - Print this message and exit.

If you only want to perform certain configuration steps, use these options instead of -A:
  -R|--install-rook           - Install Rook storage orchestrator.
  -S|--install-serverless     - Install Red Hat Openshift Serverless operator.
  -K|--setup-knative-serving  - Create a Knative Serving instance (requires OpenShift Serverless operator).
  -L|--label-node             - Add an affinity label to a worker node (required for E2E tests to work).
  -I|--create-icsp            - Add an ImageContentSourcePolicy to mirror the production repository (icr.io) to 
                                  staging repository (cp.stg.icr.io).
  -P|--add-pull-secret        - Add the pull secret provided by -k for cp.stg.icr.io to the cluster's 
                                  .dockerconfigjson (this enables pulling images from cp.stg.icr.io).
  -U|--add-cluster-admin <user>:<pass>  
                              - Add an admin user to the cluster with the given username and password to
                                  allow logging in without kubeadmin. (Adds HTPasswd secret and identity
                                  provider to the cluster.)

EOF
}

yel="\033[1;33m"
red="\033[1;31m"
grn="\033[0;32m"
blu="\033[1;36m"
gry="\033[1;30m"
end="\033[0m"

header() {
  echo
  echo -e "${grn}=============================================================================================="
  echo "$1"
  echo -e "==============================================================================================${end}"
  echo
}

stage() {
  echo -e "${blu}Stage: $1 ${end}"
}

skip() {
  echo -e "${gry} ...   $1 ${end}"
}

warn() {
  echo -e "${yel}Warning:${end} $1"
}

error() {
  echo -e "${red}Error:${end} $1"
}


main() {
  parse_args "$@"

  header "Starting cluster configuration..."

  if [[ $INSTALL_ROOK ]]; then
    stage "Install Rook storage orchestrator"
    install_rook
  else
    skip "Skipping installation of Rook storage orchestrator."
  fi

  if [[ $INSTALL_SERVERLESS ]]; then
    stage "Install Red Hat Openshift Serverless operator"
    install_serverless
  else
    skip "Skipping installation of Red Hat Openshift Serverless operator."
  fi

  if [[ $SETUP_KNATIVE_SERVING ]]; then
    stage "Create Knative Serving instance"
    setup_knative_serving
  else
    skip "Skipping creation of Knative Serving instance."
  fi

  if [[ $LABEL_NODE ]]; then
    stage "Add affinity label to a worker node"
    add_affinity_label_to_node
  else
    skip "Skipping adding affinity label to worker node."
  fi

  if [[ $CREATE_ICSP ]]; then
    stage "Add ImageContentSourcePolicy to mirror to staging repository"
    create_image_content_source_policy
  else
    skip "Skipping adding ImageContentSourcePolicy to mirror to staging repository."
  fi

  if [[ $ADD_PULL_SECRET ]]; then
    stage "Add a pull secret for cp.stg.icr.io to .dockerconfigjson"
    add_stg_registry_pull_secret
  else
    skip "Skipping adding a pull secret for cp.stg.icr.io to .dockerconfigjson."
  fi

  if [[ $ADD_CLUSTER_ADMIN ]]; then
    stage "Add a new cluster-admin to the cluster"
    if ! echo "${NEW_ADMIN_CREDS}" | grep '^.*:.*$' >/dev/null; then
      error "The credential provided to --add-cluster-admin are invalid. Please provide a username and password in the format \"<user>:<pass>\" (the colon is required)"
      exit 1
    else
      add_cluster_admin
    fi
  else
    skip "Skipping adding a new cluster-admin to the cluster."
  fi

  header "Cluster configuration complete."

}


parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      -u|--user)
        shift
        readonly CLUSTER_USERNAME="${1}"
        ;;
      -p|--pass)
        shift
        readonly CLUSTER_PASSWORD="${1}"
        ;;
      -k|--key)
        shift
        readonly REGISTRY_KEY="${1}"
        ;;
      -A|--all)
        readonly INSTALL_ROOK=true
        readonly INSTALL_SERVERLESS=true
        readonly SETUP_KNATIVE_SERVING=true
        readonly LABEL_NODE=true
        readonly CREATE_ICSP=true
        readonly ADD_PULL_SECRET=true
        ;;
      -R|--install-rook)
        readonly INSTALL_ROOK=true
        ;;
      -S|--install-serverless)
        readonly INSTALL_SERVERLESS=true
        ;;
      -K|--setup-knative-serving)
        readonly SETUP_KNATIVE_SERVING=true
        ;;
      -L|--label-node)
        readonly LABEL_NODE=true
        ;;
      -I|--create-icsp)
        readonly CREATE_ICSP=true
        ;;
      -P|--add-pull-secret)
        readonly ADD_PULL_SECRET=true
        ;;
      -U|--add-cluster-admin)
        shift
        readonly ADD_CLUSTER_ADMIN=true
        readonly NEW_ADMIN_CREDS="${1}"
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        if [[ ! -z "${CLUSTER_URL}" ]]; then
          error "Invalid argument \"$1\""
          usage
          exit 1
        fi
        readonly CLUSTER_URL="${1}"
        ;;
    esac
    shift
  done

  if [[ -z "${CLUSTER_USERNAME}" ]]; then
    warn "Cluster username not provided (-u/--user); defaulting to 'kubeadmin'."
    readonly CLUSTER_USERNAME="kubeadmin"
  fi

  if [[ $ADD_PULL_SECRET ]]; then
    if [[ -z "${REGISTRY_KEY}" ]]; then
      error "Registry key (-k) not provided; this argument is required when using -A or -P."
      usage
      exit 1
    fi
  fi

  if logged_into_cluster; then
    current_cluster_url="$(oc status | grep -o 'https://.*$')"
    if ! echo "$current_cluster_url" | grep "${CLUSTER_URL}" >/dev/null; then
      error "You're currently logged into a cluster ($current_cluster_url) other than the one you specified ($CLUSTER_URL). \
      \n       To keep your context clean, please log out of your current cluster and try again, or log into your specified cluster."
      exit 1
    fi

    current_cluster_user="$(oc whoami | sed s/://g)"
    if [[ "${CLUSTER_USERNAME:-kubeadmin}" != "$current_cluster_user" ]]; then
      error "You're currently logged into the cluster as a user ($current_cluster_user) other than the one you specified (${CLUSTER_USERNAME:-kubeadmin}). \
      \n       For this script to work best, please log out of your current cluster and try again, or switch to the correct user."
      exit 1
    fi
  else
      if [[ -z "${CLUSTER_PASSWORD}" ]]; then
        error "Cluster password not provided (-p/--pass); this argument is required."
        usage
        exit 1
      fi

      if [[ -z "${CLUSTER_URL}" ]]; then
        error "Cluster URL not provided; this argument is required."
        usage
        exit 1
      fi

      oc login "${CLUSTER_URL}" -u "${CLUSTER_USERNAME}" -p "${CLUSTER_PASSWORD}"

      if [[ $? != 0 ]]; then
        error "Unable to login to cluster ${CLUSTER_URL} with the given username and password."
        exit 1
      fi
  fi
}

logged_into_cluster() {
  oc whoami 2>&1 1&>/dev/null
}

install_rook() {
  if ! oc get storageclass | grep rook-ceph >/dev/null; then
    echo "Installing Rook storage orchestrator..."

    cur_dir="$(pwd)"
    tmp_dir=$(mktemp -d -t ceph-XXXXXXXXXX)
    cd "$tmp_dir"

    git clone --single-branch -b v1.5.11 https://github.com/rook/rook.git
    cd rook/cluster/examples/kubernetes/ceph

    oc create -f crds.yaml
    oc create -f common.yaml
    oc create -f operator-openshift.yaml
    oc create -f cluster.yaml
    oc create -f ./csi/rbd/storageclass.yaml
    oc create -f ./csi/rbd/pvc.yaml
    oc create -f ./csi/rbd/snapshotclass.yaml
    oc create -f filesystem.yaml
    oc create -f ./csi/cephfs/storageclass.yaml
    oc create -f ./csi/cephfs/pvc.yaml
    oc create -f ./csi/cephfs/snapshotclass.yaml
    oc create -f toolbox.yaml

    oc patch storageclass rook-cephfs -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
    cd "$cur_dir"
    echo
  else
    echo "Rook storage orchestrator is already installed."
  fi
}

install_serverless() {
  if ! oc get subs -n openshift-operators | grep serverless-operator >/dev/null; then
    echo "Installing Red Hat Openshift Serverless operator..."
    name="serverless-operator"
    packageManifest="$(oc get packagemanifests $name -n openshift-marketplace -o jsonpath="{.status.catalogSource},{.status.catalogSourceNamespace},{.status.channels[?(@.name=='stable')].currentCSV}")"
    catalogSource="$(echo $packageManifest | cut -d, -f1)"
    catalogSourceNamespace="$(echo $packageManifest | cut -d, -f2)"
    currentCSV="$(echo $packageManifest | cut -d, -f3)"
    cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: $name
  namespace: openshift-operators
  generateName: serverless-operator-
spec:
  channel: stable
  installPlanApproval: Automatic
  name: $name
  source: $catalogSource
  sourceNamespace: $catalogSourceNamespace
  startingCSV: $currentCSV
EOF
    echo
  else
    echo "Red Hat Openshift Serverless operator is already installed."
  fi
}

setup_knative_serving() {
  if ! oc get knativeserving.operator.knative.dev/knative-serving -n knative-serving >/dev/null; then
    if [[ $INSTALL_SERVERLESS ]]; then
      echo "Waiting 30 seconds for Serverless operator to finish being set up..."
      sleep 30
    fi
    wait_count=0
    while [ $wait_count -le 20 ]
    do
      echo "Creating Knative Serving instance..."
      cat <<EOF | oc apply -f -
apiVersion: operator.knative.dev/v1alpha1
kind: KnativeServing
metadata:
    name: knative-serving
    namespace: knative-serving
EOF
      [[ $? == 0 ]] && break
      warn "Knative Serving configuration failed (probably because the Serverless operator isn't done being set up). Trying again in 15 seconds."
      ((wait_count++))
      sleep 15
    done
    echo
  else
    echo "Knative Serving instance is already created."
  fi
}

add_affinity_label_to_node() {
  affinity_label="kuttlTest=test1"
  labeled_node="$(oc get nodes -l "$affinity_label" -o name)"
  if [[ -z "$labeled_node" ]]; then
    first_worker="$(oc get nodes | grep worker | head -1 | awk '{print $1}')"
    echo "Adding affinity label ($affinity_label) to worker node $first_worker..."
    oc label --overwrite node "$first_worker" "$affinity_label"
    echo
  else
    echo "Affinity label ($affinity_label) already exists on worker node $labeled_node."
  fi
}

create_image_content_source_policy() {
  if ! oc get imagecontentsourcepolicy | grep mirror-config >/dev/null; then
    echo "Adding ImageContentSourcePolicy to mirror to staging repository..."
    cat <<EOF | oc apply -f -
apiVersion: operator.openshift.io/v1alpha1
kind: ImageContentSourcePolicy
metadata:
    name: mirror-config
spec:
    repositoryDigestMirrors:
    - mirrors:
      - cp.stg.icr.io/cp
      source: cp.icr.io/cp
    - mirrors:
      - cp.stg.icr.io/cp
      source: icr.io/cpopen
EOF
    echo
  else
    echo "ImageContentSourcePolicy to mirror to staging repository already exists."
  fi
}

add_stg_registry_pull_secret() {
  dockerconfigjson="$(oc extract secret/pull-secret -n openshift-config --to=-)"
  if [[ "$(echo $dockerconfigjson | jq '.auths["cp.stg.icr.io"]')" == "null" ]]; then
    echo "Adding a pull secret for cp.stg.icr.io to .dockerconfigjson..."
    auth="$(echo "iamapikey:${REGISTRY_KEY}" | base64)"
    echo $dockerconfigjson | jq --arg auth "$auth" '.auths["cp.stg.icr.io"]={"email":"unused","auth":$auth}' > /tmp/.dockerconfigjson
    oc set data secret/pull-secret -n openshift-config --from-file=/tmp/.dockerconfigjson
    rm /tmp/.dockerconfigjson
  else
    echo ".dockerconfigjson already contains a pull secret for cp.stg.icr.io."
  fi
}

add_cluster_admin() {
  NEW_ADMIN_USER="$(echo "${NEW_ADMIN_CREDS}" | cut -d: -f1)"
  NEW_ADMIN_PASS="$(echo "${NEW_ADMIN_CREDS}" | cut -d: -f2)"

  cur_dir="$(pwd)"
  tmp_dir=$(mktemp -d -t htpasswd-XXXXXXXXXX)
  cd "$tmp_dir"

  if ! oc get user "${NEW_ADMIN_USER}" >/dev/null 2>&1; then
    if oc get secret htpass-secret >/dev/null 2>&1; then
      echo "Adding new user \"${NEW_ADMIN_USER}\" to cluster's HTPasswd secret..."
      oc get secret htpass-secret -ojsonpath={.data.htpasswd} -n openshift-config | base64 --decode > users.htpasswd
      htpasswd -bB users.htpasswd "${NEW_ADMIN_USER}" "${NEW_ADMIN_PASS}"

      echo "Updating HTPasswd secret on cluster..."
      oc create secret generic htpass-secret --from-file=htpasswd=users.htpasswd --dry-run=client -o yaml -n openshift-config | oc replace -f -
    else
      echo "Creating HTPasswd file for new user \"${NEW_ADMIN_USER}\"..."
      htpasswd -c -B -b users.htpasswd "${NEW_ADMIN_USER}" "${NEW_ADMIN_PASS}"

      echo "Adding HTPasswd secret to cluster..."
      oc create secret generic htpass-secret --from-file=htpasswd=users.htpasswd -n openshift-config
    fi

    cd "$cur_dir"

    echo "Add HTPasswd identity provider to cluster..."
    cat <<EOF | oc apply -f -
apiVersion: config.openshift.io/v1
kind: OAuth
metadata:
  name: cluster
spec:
  identityProviders:
  - name: htpasswd-provider
    mappingMethod: claim
    type: HTPasswd
    htpasswd:
      fileData:
        name: htpass-secret
EOF
        
    wait_count=0
    while [ $wait_count -le 20 ]
    do
      echo "Adding cluster-admin role to \"${NEW_ADMIN_USER}\" user..."
      oc adm policy add-cluster-role-to-user cluster-admin "${NEW_ADMIN_USER}"
      [[ $? == 0 ]] && break
      warn "Adding cluster-admin role to \"${NEW_ADMIN_USER}\" user failed (probably because the identity provider isn't done being set up). Trying again in 15 seconds."
      ((wait_count++))
      sleep 15
    done
    echo
  else
    echo "The user \"${NEW_ADMIN_USER}\" already exists on this cluster."
  fi
}

main "$@"
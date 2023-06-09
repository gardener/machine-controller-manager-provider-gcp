# HOW TO MANUALLY TEST CMEK FEATURE 

### Pre-Requisites

1. Install gcloud CLI: `brew install gcloud`
1. Login into gcloud: `gcloud auth login`
1. Setup a GCP shoot cluster using a service account
1. Setup Cloud KMS Key and Assign IAM role:
    1. Create a Cloud KMS keyring: `gcloud kms keyrings create <keyRingName> --location=<region>`. (This requires IAM privileges to be assigned to you)
       1. `keyRingName` should be alphanumeric
       2. `region` is a Google Cloud Location (aka region) like `europe-west1`
    2. Create a Cloud KMS Key for encryption purpose: `gcloud kms keys create alpha --keyring=<keyRingName> --location=<region> --purpose encryption`
    3. List to get fully qualified path of Cloud KMS Key: 
    ```
    gcloud kms keys list --keyring=<keyRingName> --location=<region>
    NAME                                                                                                                                                       PURPOSE
    projects/projectId/locations/region/keyRings/keyRingName/cryptoKeys/alpha    ENCRYPT_DECRYPT
    ```
    4. Give the service account used to create the shoot cluster relevant IAM role to use this key
    ```
    gcloud kms keys add-iam-policy-binding alpha --location <region>  --keyring <keyRingName> --member serviceAccount:serviceAccount:user@projectId.iam.gserviceaccount.com  --role roles/cloudkms.cryptoKeyEncrypterDecrypter
    ```
1. Checkout the `machine-controller-manager` ,`machine-controller-manager-provider-gcp` and `gardener-extension-provider-gcp` repos in peer directories under `$GOPATH/src/github.com/gardener`.
1. Ensure docker VM and docker CLI is installed on your box and configure docker for building amd64 images using `docker buildx create --name amdlinux--use --bootstrap`
   1. See: https://docs.docker.com/engine/reference/commandline/buildx_create/

### MCM GCP Test

1. Setup and Start Local MCM: 
     1. `cd $GOPATH/src/github.com/gardener/machine-controller-manager`
     2. `make download-kubeconfigs`
     4. `make start 2>&1 | tee /tmp/mcm.log`
1. Start Local MC 
     1. `cd $GOPATH/src/github.com/gardener/machine-controller-manager-provider-gcp`
     2. `make start 2>&1 | tee /tmp/mcm-gcp.log`
     3. Wait for any pending reconciles to finish.
1. Edit the `MachineClass` yaml
     1. `k edit mcc`
     2.   Modify `providerSpec.disks` as follows: adding `kmsKeyName` and `kmsKeyServiceAccount` as illustrated below (example below makes boot disk CMEK encrypted):
     ```yaml
          - autoDelete: true
             boot: true
             encryption:
               kmsKeyName: projects/projectId/locations/<region>/keyRings/<keyRingName>/cryptoKeys/alpha
               kmsKeyServiceAccount: user@projectId.iam.gserviceaccount.com
      ```
      Please note that you can omit `kmsKeyServiceAccount`. This will then default to the Compute Engine Service Agent Account. See https://cloud.google.com/iam/docs/service-agents#compute-engine-service-agent
1.  Trigger restart of `Machine`
      1. List the existing `Running` machine:
       ```
      k get machine
      shoot--userid--shootName-worker-alpha-z1-9485d-xprld
       ```
    1. Add the `force-deletion` label to the machine. (So restart is faster)
        ```
           k label mc shoot--userid--shootName-worker-alpha-z1-9485d-xprld force-deletion=true
        ```
     1. Delete the machine: `k delete mc shoot--userid--shootName-worker-alpha-z1-9485d-xprld`
     1. Wait for reconcile to finish and confirm new machine is `Running`
       ```
       k get machine                                    
       NAME                                                        STATUS    AGE   
       shoot--userid--shootName-worker-alpha-z1-9485d-96fkl   Running   67m   
       ```
 1. Confirm Disk Creation/Encryption.
       1. Test to see that corresponding boot disk is created via gcloud CLI. (The boot disk name is same as machine name)
          ```
             gcloud compute disks list | grep 96fkl

             shoot--userid---shootName-worker-alpha-z1-9485d-96fkl  <zoneName> 50       pd-balanced  READY
           ```
       1. Confirm that disk is using CMEK encryption key:
       ```
           gcloud compute disks describe shoot--shootName-worker-alpha-z1-9485d-96fkl --zone <zoneName> 
           
           # Output Trimmed
           ...
           diskEncryptionKey:
               kmsKeyName: projects/projectId/locations/<region>/keyRings/<keyRingName>/cryptoKeys/alpha/cryptoKeyVersions/<revision>
               kmsKeyServiceAccount: user@projectId.iam.gserviceaccount.com
            id: '843731016874110129'
           ...
       ```

### Gardener Shoot Test

1. Setup a local Gardener shoot cluster. One way to do this is by following instructions at https://github.com/gardener/gardener/blob/master/docs/development/local_setup.md and ensure one installs the GCP extension provider. 
1. Build and push the images for `garden-extension-provider-gcp` and `machine-controller-manager-provider-gcp` to your docker repo. 
   1. Inside dir `gardener-extension-provider-gcp`, execute: 
      - `docker buildx build --push --platform linux/amd64 --tag dockeruser/extension-provider-gcp:latest --target gardener-extension-provider-gcp .`
   1. In dir: `machine-controller-manager-provider-gcp`, execute:    
       - `docker buildx build --push --platform linux/amd64 --tag dockeruser/machine-controller-manager-provider-gcp:latest --target machine-controller` 
1. Change the `ControllerDeployment` yaml for GCP extension provider name: `provider-gcp` by modifying the repository and SHA tag for its image as well as the dependent image of `machine-controller-manager-provider-gcp` to the images pushed above. Apply the controller deployment to the nodeless local garden.
1. Change the shoot yaml by modifying the `worker` volume and `providerConfig` section as below
      ```yaml
         workers:
            - cri:
               name: containerd
            name: worker-1
            # ... skipped
            volume:
               type: pd-balanced
               size: 50Gi
            dataVolumes:
               - name: alpha
                  type: pd-balanced
                  size: 25Gi
               - name: beta
                  type: SCRATCH
                  size: 25Gi
            zones:
               - europe-west1-c
            systemComponents:
               allow: true
            providerConfig:
               apiVersion: gcp.provider.extensions.gardener.cloud/v1alpha1
               kind: WorkerConfig
               volume:
                  interface: "SCSI" # applies only for scratch disks. misleading field name
                  encryption:
                     kmsKeyName: "projects/projectId/locations/<zoneName>/keyRings/<keyRingName>/cryptoKeys/alpha"
                     kmsKeyServiceAccount: "user@projectId.iam.gserviceaccount.com" 
      ```
1. Apply the shoot YAML against the nodeless cluster and wait until the shoot has reconciled and the machine(s) in the worker pool are in `Running` phase.
1. Use `gardenctl` to target the shooted seed and get the provider machine name and zone
   ```
   kubectl get mc -A -o=jsonpath='{.items[*].spec.providerID}'
   gce:///projectId/zone/shoot--userid--shootName-worker-1-z1-65c6b-9ptm7
   ```
1. You can check that the encrypted disks have been created:
    1. Describe the instance: 
    `gcloud compute instances describe shoot--userid--shootName-worker-1-z1-65c6b-9ptm7 --zone zone` 
    1. The persistent disks should look like the one below with a disk encryption key.
    ```yaml
      disks:
      - autoDelete: true
      boot: true
      deviceName: persistent-disk-0
      diskEncryptionKey:
         kmsKeyName: projects/projectId/locations/<zoneName>/keyRings/<keyRingName>/cryptoKeys/alpha"
         kmsKeyServiceAccount: user@projectId.iam.gserviceaccount.com 
      diskSizeGb: '50'
      index: 0
      interface: SCSI
      kind: compute#attachedDisk
      mode: READ_WRITE
      source: <...>
      type: PERSISTENT
      - autoDelete: true
      boot: false
      deviceName: local-ssd-0
      diskSizeGb: '375'
      index: 1
      interface: SCSI
      kind: compute#attachedDisk
      mode: READ_WRITE
      savedState: DISK_SAVED_STATE_UNSPECIFIED
      type: SCRATCH
      - autoDelete: true
      boot: false
      deviceName: persistent-disk-2
      diskEncryptionKey:
         kmsKeyName: projects/projectId/locations/<zoneName>/keyRings/<keyRingName>/cryptoKeys/alpha
         kmsKeyServiceAccount: user@projectId.iam.gserviceaccount.com 
      diskSizeGb: '25'
      index: 2
      interface: SCSI
      kind: compute#attachedDisk
      mode: READ_WRITE
      source: <...>
      type: PERSISTENT
    ```
1. One can change the shoot YAML by adding/removing volumes. Once the shoot has reconciled with the new machines in the Running phase, the persistent disks associated with the VM should be added/removed and still be configured with the specified encryption key. 

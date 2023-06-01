# HOW TO MANUALLY TEST CMEK FEATURE 

1. Setup a GCP shoot cluster using a service account
1. Setup Cloud KMS Key and Assign IAM role:
    1. Install gcloud CLI: `brew install gcloud`
    1. Login into gcloud: `gcloud auth login`
    1. Create a Cloud KMS keyring: `gcloud kms keyrings create cmektest --location=us-east4`. (This requires IAM privileges to be assigned to you)
    1. Create a Cloud KMS Key for encryption purpose: `gcloud kms keys create alpha --keyring=cmektest --location=us-east4 --purpose encryption`
    1. List to get fully qualified path of Cloud KMS Key: 
    ```
    gcloud kms keys list --keyring=cmektest --location=us-east4
    NAME                                                                                                                                                       PURPOSE
    projects/projectId/locations/us-east4/keyRings/cmektest/cryptoKeys/alpha    ENCRYPT_DECRYPT
    ```
    1. Give the service account used to create the shoot cluster relevant IAM role to use this key
    ```
    gcloud projects add-iam-policy-binding projectId --member serviceAccount:user@projectId.iam.gserviceaccount.com --role roles/cloudkms.cryptoKeyEncrypterDecrypter
    ```
1. Checkout the `machine-controller-manager` and `machine-controller-manager-provider-gcp` in peer directories under `$GOPATH/src/github.com/gardener`.
3. Setup and Start Local MCM: 
     1. `cd $GOPATH/src/github.com/gardener/machine-controller-manager`
     2. `make download-kubeconfigs`
     4. `make start 2>&1 | tee /tmp/mcm.log`
5. Start Local MC 
     1. `cd $GOPATH/src/github.com/gardener/machine-controller-manager-provider-gcp`
     2. `make start 2>&1 | tee /tmp/mcm-gcp.log`
     3. Wait for any pending reconciles to finish.
6. Edit the `MachineClass` yaml
     1. `k edit mcc`
     2.   Modify `providerSpec.disks` as follows: adding `kmsKeyName` and `kmsKeyServiceAccount` as illustrated below (example below makes boot disk CMEK encrypted):
     ```yaml
          - autoDelete: true
             boot: true
             kmsKeyName: projects/projectId/locations/us-east4/keyRings/cmektest/cryptoKeys/alpha
             kmsKeyServiceAccount: user@projectId.iam.gserviceaccount.com
      ```
      Please note that you can omit `kmsKeyServiceAccount`. This will then default to the Compute Engine Service Agent Account. See https://cloud.google.com/iam/docs/service-agents#compute-engine-service-agent
7.  Trigger restart of `Machine`
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
 11. Confirm Disk Creation/Encryption.
       1. Test to see that corresponding boot disk is created via gcloud CLI. (The boot disk name is same as machine name)
          ```
             gcloud compute disks list | grep 96fkl
             shoot--userid---shootName-worker-alpha-z1-9485d-96fkl  us-east4-c      zone            50       pd-balanced  READY
           ```
       1. Confirm that disk is using CMEK encryption key:
       ```
           gcloud compute disks describe shoot--shootName-worker-alpha-z1-9485d-96fkl --zone us-east4-c worker-alpha-z1-9485d-96fkl --zone us-east4-c
           creationTimestamp: '2023-05-25T00:46:41.136-07:00'
           diskEncryptionKey:
               kmsKeyName: projects/projectId/locations/us-east4/keyRings/cmektest/cryptoKeys/alpha/cryptoKeyVersions/2
               kmsKeyServiceAccount: user@projectId.iam.gserviceaccount.com
            id: '843731016874110129'
       ```

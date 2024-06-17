## Specification
### ProviderSpec Schema
<br>
<h3 id="settings.gardener.cloud/v1alpha1.GCPProviderSpec">
<b>GCPProviderSpec</b>
</h3>
<p>
<p>GCPProviderSpec contains the fields of
provider spec that the plugin expects</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Type</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code>
</td>
<td>
string
</td>
<td>
<code>
settings.gardener.cloud/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code>
</td>
<td>
string
</td>
<td>
<code>GCPProviderSpec</code>
</td>
</tr>
<tr>
<td>
<code>APIVersion</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>APIVersion refers to the APIVersion of the object</p>
</td>
</tr>
<tr>
<td>
<code>canIpForward</code>
</td>
<td>
<em>
bool
</em>
</td>
<td>
<p>CanIpForward: Allows this instance to send and receive packets with
non-matching destination or source IPs. This is required if you plan
to use this instance to forward routes. For more information, see
Enabling IP Forwarding.</p>
</td>
</tr>
<tr>
<td>
<code>deletionProtection</code>
</td>
<td>
<em>
bool
</em>
</td>
<td>
<p>DeletionProtection: Whether the resource should be protected against
deletion.</p>
</td>
</tr>
<tr>
<td>
<code>description</code>
</td>
<td>
<em>
*string
</em>
</td>
<td>
<p>Description: An optional description of this resource. Provide this
property when you create the resource.</p>
</td>
</tr>
<tr>
<td>
<code>disks</code>
</td>
<td>
<em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPDisk">
[]GCPDisk
</a>
</em>
</td>
<td>
<p>Disks: Array of disks associated with this instance. Persistent disks
must be created before you can assign them.</p>
</td>
</tr>
<tr>
<td>
<code>gpu</code>
</td>
<td>
<em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPGpu">
GCPGpu
</a>
</em>
</td>
<td>
<p>Gpu: Configurations related to GPU which would be attached to the instance. Enough
Quota of the particular GPU should be available.</p>
</td>
</tr>
<tr>
<td>
<code>labels</code>
</td>
<td>
<em>
map[string]string
</em>
</td>
<td>
<p>Labels: Labels to apply to this instance.</p>
</td>
</tr>
<tr>
<td>
<code>machineType</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>MachineType: Full or partial URL of the machine type resource to use
for this instance, in the format:
zones/zone/machineTypes/machine-type. This is provided by the client
when the instance is created. For example, the following is a valid
partial url to a predefined machine
type:
zones/us-central1-f/machineTypes/n1-standard-1</p>
<p>To create a custom machine type, provide a URL to a machine type in
the following format, where CPUS is 1 or an even number up to 32 (2,
4, 6, &hellip; 24, etc), and MEMORY is the total memory for this instance.
Memory must be a multiple of 256 MB and must be supplied in MB (e.g.
5 GB of memory is 5120
MB):
zones/zone/machineTypes/custom-CPUS-MEMORY</p>
<p>For example: zones/us-central1-f/machineTypes/custom-4-5120</p>
<p>For a full list of restrictions, read the Specifications for custom
machine types.</p>
</td>
</tr>
<tr>
<td>
<code>metadata</code>
</td>
<td>
<em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPMetadata">
[]GCPMetadata
</a>
</em>
</td>
<td>
<p>Metadata: The metadata key/value pairs assigned to this instance.
This includes custom metadata and predefined keys.</p>
</td>
</tr>
<tr>
<td>
<code>minCpuPlatform</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>MinCpuPlatform: The name of the minimum CPU platform that is requested
for this instance.</p>
</td>
</tr>
<tr>
<td>
<code>networkInterfaces</code>
</td>
<td>
<em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPNetworkInterface">
[]GCPNetworkInterface
</a>
</em>
</td>
<td>
<p>NetworkInterfaces: An array of network configurations for this
instance. These specify how interfaces are configured to interact
with other network services, such as connecting to the internet.
Multiple interfaces are supported per instance.</p>
</td>
</tr>
<tr>
<td>
<code>region</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Region: in which instance is to be deployed</p>
</td>
</tr>
<tr>
<td>
<code>scheduling</code>
</td>
<td>
<em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPScheduling">
GCPScheduling
</a>
</em>
</td>
<td>
<p>Scheduling: Sets the scheduling options for this instance.</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccounts</code>
</td>
<td>
<em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPServiceAccount">
[]GCPServiceAccount
</a>
</em>
</td>
<td>
<p>ServiceAccounts: A list of service accounts, with their specified
scopes, authorized for this instance. Only one service account per VM
instance is supported.</p>
<p>Service accounts generate access tokens that can be accessed through
the metadata server and used to authenticate applications on the
instance. See Service Accounts for more information.</p>
</td>
</tr>
<tr>
<td>
<code>tags</code>
</td>
<td>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Tags: to be placed on the VM</p>
</td>
</tr>
<tr>
<td>
<code>zone</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Zone: in which instance is to be deployed</p>
</td>
</tr>
</tbody>
</table>
<br>
<h3 id="settings.gardener.cloud/v1alpha1.GCPDisk">
<b>GCPDisk</b>
</h3>
<p>
(<em>Appears on:</em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPProviderSpec">GCPProviderSpec</a>)
</p>
<p>
<p>GCPDisk describes disks for GCP.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Type</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>autoDelete</code>
</td>
<td>
<em>
*bool
</em>
</td>
<td>
<p>AutoDelete: Specifies whether the disk will be auto-deleted when the
instance is deleted (but not when the disk is detached from the
instance).</p>
</td>
</tr>
<tr>
<td>
<code>boot</code>
</td>
<td>
<em>
bool
</em>
</td>
<td>
<p>Boot: Indicates that this is a boot disk. The virtual machine will
use the first partition of the disk for its root filesystem.</p>
</td>
</tr>
<tr>
<td>
<code>sizeGb</code>
</td>
<td>
<em>
int64
</em>
</td>
<td>
<p>SizeGb: Specifies the size of the disk in base-2 GB.</p>
</td>
</tr>
<tr>
<td>
<code>type</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Type: Specifies the disk type to use to create the instance. If
not specified, the default is pd-standard, specified using the full
URL. For
example:
<a href="https://www.googleapis.com/compute/v1/projects/project/zones/">https://www.googleapis.com/compute/v1/projects/project/zones/</a>
zone/diskTypes/pd-standard</p>
<p>Other values include pd-ssd and local-ssd. If you define this field,
you can provide either the full or partial URL. For example, the
following are valid values:
- <a href="https://www.googleapis.com/compute/v1/projects/project/zones/zone/diskTypes/diskType">https://www.googleapis.com/compute/v1/projects/project/zones/zone/diskTypes/diskType</a>
- projects/project/zones/zone/diskTypes/diskType
- zones/zone/diskTypes/diskType  Note that for InstanceTemplate, this
is the name of the disk type, not URL.
If you use &ldquo;SCRATCH&rdquo; as the value, it is defaulted to local-ssd</p>
</td>
</tr>
<tr>
<td>
<code>interface</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Interface: Specifies the disk interface to use for attaching this
disk, which is either SCSI or NVME. The default is SCSI. Persistent
disks must always use SCSI and the request will fail if you attempt
to attach a persistent disk in any other format than SCSI. Local SSDs
can use either NVME or SCSI. For performance characteristics of SCSI
over NVMe, see Local SSD performance.</p>
<p>Possible values:
&ldquo;NVME&rdquo;
&ldquo;SCSI&rdquo;
This is only applied when the disk type is &ldquo;SCRATCH&rdquo; currently</p>
</td>
</tr>
<tr>
<td>
<code>image</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Image: The source image to create this disk. When creating a
new instance, one of initializeParams.sourceImage or disks.source is
required except for local SSD.</p>
<p>To create a disk with one of the public operating system images,
specify the image by its family name. For example, specify
family/debian-9 to use the latest Debian 9
image:
projects/debian-cloud/global/images/family/debian-9</p>
<p>Alternatively, use a specific version of a public operating system
image:
projects/debian-cloud/global/images/debian-9-stretch-vYYYYMMDD</p>
<p>To create a disk with a custom image that you created, specify the
image name in the following
format:
global/images/my-custom-image</p>
<p>You can also specify a custom image by its image family, which
returns the latest version of the image in that family. Replace the
image name with
family/family-name:
global/images/family/my-image-family</p>
<p>If the source image is deleted later, this field will not be set.</p>
</td>
</tr>
<tr>
<td>
<code>encryption</code>
</td>
<td>
<em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPDiskEncryption">
GCPDiskEncryption
</a>
</em>
</td>
<td>
<p>Encryption: Encryption details for this disk</p>
</td>
</tr>
<tr>
<td>
<code>labels</code>
</td>
<td>
<em>
map[string]string
</em>
</td>
<td>
<p>Labels: Labels to apply to this disk. These can be later modified by
the disks.setLabels method. This field is only applicable for
persistent disks.</p>
</td>
</tr>
<tr>
<td>
<code>provisionedIops</code>
</td>
<td>
<em>
int64
</em>
</td>
<td>
<p>ProvisionedIops of disk to create.
Only for use with disks of type pd-extreme and hyperdisk-extreme.
The IOPS must be specified within defined limits
the value zero will be omitted from the request because GCP client
will not write any &ldquo;empty&rdquo; values to the request</p>
</td>
</tr>
<tr>
<td>
<code>provisionedThroughput</code>
</td>
<td>
<em>
int64
</em>
</td>
<td>
<p>ProvisionedThroughput of disk to create.
Only for hyperdisk-balanced or hyperdisk-throughput volumes,
measured in MiB per second, that the disk can handle.
The throughput must be specified within defined limits
the value zero will be omitted from the request because GCP client
will not write any &ldquo;empty&rdquo; values to the request</p>
</td>
</tr>
</tbody>
</table>
<br>
<h3 id="settings.gardener.cloud/v1alpha1.GCPDiskEncryption">
<b>GCPDiskEncryption</b>
</h3>
<p>
(<em>Appears on:</em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPDisk">GCPDisk</a>)
</p>
<p>
<p>GCPDiskEncryption holds references to encryption data</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Type</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>kmsKeyName</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>KmsKeyName: key name of the cloud kms disk encryption key. Not optional</p>
</td>
</tr>
<tr>
<td>
<code>kmsKeyServiceAccount</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>KmsKeyServiceAccount: The service account granted the <code>roles/cloudkms.cryptoKeyEncrypterDecrypter</code> for the key name.
If empty, then the role should be given to the Compute Engine Service Agent Account. This usually has the format
service-PROJECT_NUMBER@compute-system.iam.gserviceaccount.com. See: <a href="https://cloud.google.com/iam/docs/service-agents#compute-engine-service-agent">https://cloud.google.com/iam/docs/service-agents#compute-engine-service-agent</a>
One can add IAM roles using the gcloud CLI:
gcloud projects add-iam-policy-binding projectId &ndash;member
serviceAccount:name@projectIdgserviceaccount.com &ndash;role roles/cloudkms.cryptoKeyEncrypterDecrypter</p>
</td>
</tr>
</tbody>
</table>
<br>
<h3 id="settings.gardener.cloud/v1alpha1.GCPGpu">
<b>GCPGpu</b>
</h3>
<p>
(<em>Appears on:</em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPProviderSpec">GCPProviderSpec</a>)
</p>
<p>
<p>GCPGpu describes gpu configurations for GCP</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Type</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>acceleratorType</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>count</code>
</td>
<td>
<em>
int64
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<br>
<h3 id="settings.gardener.cloud/v1alpha1.GCPMetadata">
<b>GCPMetadata</b>
</h3>
<p>
(<em>Appears on:</em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPProviderSpec">GCPProviderSpec</a>)
</p>
<p>
<p>GCPMetadata describes metadata for GCP.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Type</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>key</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Key: Key for the metadata entry. Keys must conform to the following
regexp: [a-zA-Z0-9-_]+, and be less than 128 bytes in length. This is
reflected as part of a URL in the metadata server. Additionally, to
avoid ambiguity, keys must not conflict with any other metadata keys
for the project.</p>
</td>
</tr>
<tr>
<td>
<code>value</code>
</td>
<td>
<em>
*string
</em>
</td>
<td>
<p>Value: Value for the metadata entry. These are free-form strings, and
only have meaning as interpreted by the image running in the
instance. The only restriction placed on values is that their size
must be less than or equal to 262144 bytes (256 KiB).</p>
</td>
</tr>
</tbody>
</table>
<br>
<h3 id="settings.gardener.cloud/v1alpha1.GCPNetworkInterface">
<b>GCPNetworkInterface</b>
</h3>
<p>
(<em>Appears on:</em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPProviderSpec">GCPProviderSpec</a>)
</p>
<p>
<p>GCPNetworkInterface describes network interfaces for GCP</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Type</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>disableExternalIP</code>
</td>
<td>
<em>
bool
</em>
</td>
<td>
<p>DisableExternalIP: is false, implies Attach an external IP to VM</p>
</td>
</tr>
<tr>
<td>
<code>network</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Network: URL of the network resource for this instance. When creating
an instance, if neither the network nor the subnetwork is specified,
the default network global/networks/default is used; if the network
is not specified but the subnetwork is specified, the network is
inferred.</p>
<p>This field is optional when creating a firewall rule. If not
specified when creating a firewall rule, the default network
global/networks/default is used.</p>
<p>If you specify this property, you can specify the network as a full
or partial URL. For example, the following are all valid URLs:
- <a href="https://www.googleapis.com/compute/v1/projects/project/global/networks/network">https://www.googleapis.com/compute/v1/projects/project/global/networks/network</a>
- projects/project/global/networks/network
- global/networks/default</p>
</td>
</tr>
<tr>
<td>
<code>subnetwork</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Subnetwork: The URL of the Subnetwork resource for this instance. If
the network resource is in legacy mode, do not provide this property.
If the network is in auto subnet mode, providing the subnetwork is
optional. If the network is in custom subnet mode, then this field
should be specified. If you specify this property, you can specify
the subnetwork as a full or partial URL. For example, the following
are all valid URLs:
- <a href="https://www.googleapis.com/compute/v1/projects/project/regions/region/subnetworks/subnetwork">https://www.googleapis.com/compute/v1/projects/project/regions/region/subnetworks/subnetwork</a>
- regions/region/subnetworks/subnetwork</p>
</td>
</tr>
</tbody>
</table>
<br>
<h3 id="settings.gardener.cloud/v1alpha1.GCPScheduling">
<b>GCPScheduling</b>
</h3>
<p>
(<em>Appears on:</em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPProviderSpec">GCPProviderSpec</a>)
</p>
<p>
<p>GCPScheduling describes scheduling configuration for GCP.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Type</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>automaticRestart</code>
</td>
<td>
<em>
bool
</em>
</td>
<td>
<p>AutomaticRestart: Specifies whether the instance should be
automatically restarted if it is terminated by Compute Engine (not
terminated by a user). You can only set the automatic restart option
for standard instances. Preemptible instances cannot be automatically
restarted.</p>
<p>By default, this is set to true so an instance is automatically
restarted if it is terminated by Compute Engine.</p>
</td>
</tr>
<tr>
<td>
<code>onHostMaintenance</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>OnHostMaintenance: Defines the maintenance behavior for this
instance. For standard instances, the default behavior is MIGRATE.
For preemptible instances, the default and only possible behavior is
TERMINATE. For more information, see Setting Instance Scheduling
Options.</p>
<p>Possible values:
&ldquo;MIGRATE&rdquo;
&ldquo;TERMINATE&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>preemptible</code>
</td>
<td>
<em>
bool
</em>
</td>
<td>
<p>Preemptible: Defines whether the instance is preemptible. This can
only be set during instance creation, it cannot be set or changed
after the instance has been created.</p>
</td>
</tr>
</tbody>
</table>
<br>
<h3 id="settings.gardener.cloud/v1alpha1.GCPServiceAccount">
<b>GCPServiceAccount</b>
</h3>
<p>
(<em>Appears on:</em>
<a href="#%23settings.gardener.cloud%2fv1alpha1.GCPProviderSpec">GCPProviderSpec</a>)
</p>
<p>
<p>GCPServiceAccount describes service accounts for GCP.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Type</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>email</code>
</td>
<td>
<em>
string
</em>
</td>
<td>
<p>Email: Email address of the service account.</p>
</td>
</tr>
<tr>
<td>
<code>scopes</code>
</td>
<td>
<em>
[]string
</em>
</td>
<td>
<p>Scopes: The list of scopes to be made available for this service
account.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>

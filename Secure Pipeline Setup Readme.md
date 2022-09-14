# Enabling the Multiscanner and Advanced Issue Management from a CI pipeline

The following script routines enable CI pipelines to invoke the Secure Scanner pipeline and perform the ehanced
capabilities of the MultiScanner (e.g. aqua, twistlock, va and other pluggable scanners), along with deduplication of issues across scanner types. 

Additionally the Advance Issue Management feature will be enabled, allowing for aggregation/categorization of issues by image name, CVE, squad etc.

# Required new Environment Properties for CI Pipeline

<li><code>security-scanning-pipeline-trigger</code>- the name of the trigger for running the Secure Pipeline.  This must be predefined by the admin of the Secure Scanner Pipeline</li>

<li><code>secscan_toolchain_region</code>- e.g. "us-south"</li>  

<li><code>security-scanning-pipeline-id</code>- pipeline ID derived from the url of the secure pipeline, i.e. the uuid after tekton/, e.g.
https://cloud.ibm.com/devops/pipelines/tekton/b3a9510c-87f2-4f43-9a97-59288e410906?env_id=ibm:yp:us-south
in this case  b3a9510c-87f2-4f43-9a97-59288e410906 is the security-scanning-pipeline-id</li>   

# Relevant Scripts

<li><code>ci_to_secure_pipeline_scan.sh</code>- script to call out from your CI pipeline to the Secure Pipeline to perform MultiScanning of container images and organize the results and issues with the Advanced Issue Management feature</li> 

<li><code>aim_update_yaml.sh</code>- script invoked from the Secure Pipeline setup stage to process an argument list of container images to scan for Multi Scan and Advanced Issue Management Processing</li>  

# Script Invocation

The following scripts require no parameters, but do require additional environment properties as mentioned above

<li><code>ci_to_secure_pipeline_scan.sh</code>- specify in scan-artifacts stage of the CI pipeline (no parameters) e.g. ./scripts/pipeline/ci_to_secure_pipeline_scan.sh</li> 

<li><code>aim_update_yaml.sh</code>- specify in setup stage of the Secure Pipeline (no parameters), e.g. ./aim_update_yaml.sh</li> 

# References

https://github.ibm.com/CICD-CPP/security-scans-toolchain/blob/master/SecurityScanningPipeline.md
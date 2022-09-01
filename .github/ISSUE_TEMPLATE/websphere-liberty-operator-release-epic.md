---
name: WebSphere Liberty Operator Release Epic
about: This template populates the release process steps
title: ''
labels: Epic, zenhub-dev
assignees: ''

---

- [ ] Enable and verify the upgrade support for this release version (Using catalog image produced by One Pipeline)
- [ ] Scans(Done for each release build): 
  - [ ] VA
  - [ ] Twistlock
  - [ ] SonarQube
  - [ ] Mend
  - [ ] Aquascan
- [ ] Linter
- [ ] Certification Related Work Items
- [ ] Update Liberty Sample app ([here](https://github.com/WASdev/websphere-liberty-operator/blob/1437996159871dd52d23372ffa08ff1e7eec3010/config/samples/liberty.websphere.ibm.com_v1_webspherelibertyapplications.yaml#L11)). (To use as part of early verification)
- [ ] Identify SVT Testing that is needed
- [ ] Create a total of 2 pre-release drivers in GH following these [instructions](https://github.ibm.com/websphere/operators/wiki/Creating-Operator-Releases-and-Tagging-them-for-use-with-Case#for-release-candidates-for-10x-and-above)
- [ ] Update websphereliberty-app-crd.yaml if necessary
  - [ ] Creating the Pre-Release: 
  - [ ] Kick off build using the pre-release tags
  - [ ] Set the flag to run e2e in all modes
  - [ ] Check e2e results
  - [ ] Provide image details to Mary (Out of containerized step)
- [ ] Update Liberty Sample Version (Done as close to feature complete as possible) 
- [ ] Provide customer installable code to PTC for Open Source Clearance
- [ ] Preparing for GM release in GH (Done after all changes are in and SVT has completed)
  - [ ] Create a GM Release in GH following these [instructions](https://github.ibm.com/websphere/operators/wiki/Creating-Operator-Releases-and-Tagging-them-for-use-with-Case#for-release-candidates-for-10x-and-above)
  - [ ] Kick off build using the GM release tag. **Important** Make sure that "Release" is set to true. It is "false" by default which is fine for pre-release activity
  - [ ] Set the flag to run e2e in all modes
  - [ ] Check e2e results
  - [ ] Provide image details to Mary ([Out of containerized step](https://github.ibm.com/websphere/operators/wiki/Running-the-CD-Pipeline-for-the-GM-Operator-Release))
  - [ ] Update webspehreliberty-app-crd.yaml to point to GA release image
- [ ] Push the images to the production repo (Using CD Instructions)
- [ ] Certification: create a new cert item in the tool as well as the GH repo, answer any questions
- [ ] Work with ID to refresh upgrade instructions (use WebSphere Automation Operator docs)
- [ ] Publish assessment (CASE will get published when this happens)

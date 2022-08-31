# remote-installer

The Remote Installer Operator is used to install services / packages from a remote source. The only currently supported source is 
a URL

## Description

While provisioning or maintaining Kubernetes clusters, occasionally someone needs to run `kubectl apply github.com/some/cool/yaml/files`. 
This doesn't play nicely with GitOps and similar deployment mechanisms. Using this Operator, you can define the deployment itself, as well 
as several URLs that need to be installed, and they'll be automatically taken care of

## Getting Started

<!-- TODO: Add steps for installing the actual operator --> 

<!-- TODO: Add examples for installing from a remote URL -->

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.


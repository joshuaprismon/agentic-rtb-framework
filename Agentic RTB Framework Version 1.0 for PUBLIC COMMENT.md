	

# ![LogoDescription automatically generated][image1]

# 

# **Agentic RTB Framework**

## A specification for using agent-driven containers in OpenRTB and Digital Advertising

**Version 1.0**  
*Released November 12, 2025*

		 	 	 		  
Please email support@iabtechlab.com for public comments and questions. This document is available online at [https://iabtechlab.com/standards/artf/](https://iabtechlab.com/standards/artf/)  
					  
				

			  
**© 2025 IAB Technology Laboratory** 

# About this document {#about-this-document}

The Agentic RTB Framework specification defines a foundation for implementing agent services which operate within a host platform and that the orchestrating platform can call directly to accomplish a shared goal. The model leverages containers which are deployed into the infrastructure of a host to enable delegation of critical aspects of bidstream processing to service agents in a consistent manner, with minimal cost, latency and operational impacts. The framework enables this by establishing standard requirements for container runtime behavior and by defining an API which enables reliable, protected and private bidstream mutation.

With this approach, service providers package their offering once and deploy it to any standard-compliant platform, which means they are able to focus on their unique value proposition while offloading operational concerns and scaling to the host platforms. It also creates new possibilities for innovation because host platforms maintain control of data and SLAs and therefore can provide greater access to data and more interaction opportunities to service agents without concerns about leakage, misappropriation or latency.

The Agentic RTB Framework provides significant value to host platforms which are able to “drop-in” new capabilities with minimal integration overhead and compose bid processing pipelines configured specifically for their target use cases. The standard also enables platforms to adapt quickly to changing market demands by simply adding, updating and removing components as needed, all while maintaining control of operational costs and requirements and without incurring significant integration overhead and cost.

While this approach is agentic in nature, the primary focus here is on systematic agentic integration (service to service integration), but autonomic agentic functionality (model to service) is also envisioned as part of this specification as the integrating technology matures.

There are many use cases; for example identity resolution provided by an agent, deal or segmentation activation, fraud detection pre-impression etc. This list of use cases is not fully enumerated here, because part of the goal of this specification is to allow new use cases to be integrated into the bid stream via the agentic framework. 

The specification aims to provide a general framework and best practices for deploying and operating these agents. It is not limited to a predefined set of use cases. Each use case is meant to be supported via an “intent” using the specification.

Additionally, the specification describes a standard interface with which containers can be managed using AI agents. This enables sub millisecond real-time bidding operations driven by agentic systems.

This document is primarily for technical audiences, in particular engineers and product managers wishing to implement products and features which can be solved with hosted containers and AI agents. The key takeaways for readers are:

- Understanding why to use containers and AI agents for certain use cases  
- Learning what a standard container deployment involves including digital formats for describing the container, requirements, and functions  
- Learning how to declare container capabilities and service definitions  
- Understanding example use cases and workflows  
- Recommendations of best practices for facilitating adoption across the industry

This document is developed by the IAB Tech Lab [Container Project Task Force](https://iabtechlab.com/working-groups/container-project-working-group/) which is a subgroup of the [Programmatic Supply Chain Working Group](https://iabtechlab.com/working-groups/programmatic-supply-chain-working-group/).

**License**  
Agentic RTB Framework document is licensed under a [Creative Commons Attribution 3.0 License](http://creativecommons.org/licenses/by/3.0/). To view a copy of this license, visit [creativecommons.org/licenses/by/3.0/](http://creativecommons.org/licenses/by/3.0/) or write to Creative Commons, 171 Second Street, Suite 300, San Francisco, CA 94105, USA.

**Significant Contributors**  
Joshua Prismon, *Index Exchange*; Joe Wilson, *Chalice*; Arpad Miklos, *The Trade Desk*; Brian May, *Individual*; Roni Gordon, *Index Exchange*; Ran Li, *Index Exchange*; Ben White, *OpenX*

**IAB Tech Lab Leads**  
Miguel Morales, Director Addressability & Privacy Enhancing Technologies (PETs)  
Shailley Singh, EVP Product & COO

**About IAB Tech Lab**  
The IAB Technology Laboratory is a nonprofit research and development consortium charged with producing and helping companies implement global industry technical standards and solutions. The goal of the Tech Lab is to reduce friction associated with the digital advertising and marketing supply chain while contributing to the safe growth of an industry.

The IAB Tech Lab spearheads the development of technical standards, creates and maintains a code library to assist in rapid, cost-effective implementation of IAB standards, and establishes a test platform for companies to evaluate the compatibility of their technology solutions with IAB standards, which for 18 years have been the foundation for interoperability and profitable growth in the digital advertising supply chain. Further details about the IAB Technology Lab can be found at [https://iabtechlab.com](https://iabtechlab.com).

**Disclaimer**  
THE STANDARDS, THE SPECIFICATIONS, THE MEASUREMENT GUIDELINES, AND ANY OTHER MATERIALS OR SERVICES PROVIDED TO OR USED BY YOU HEREUNDER (THE “PRODUCTS AND SERVICES”) ARE PROVIDED “AS IS” AND “AS AVAILABLE,” AND IAB TECHNOLOGY LABORATORY, INC. (“TECH LAB”) MAKES NO WARRANTY WITH RESPECT TO THE SAME AND HEREBY DISCLAIMS ANY AND ALL EXPRESS, IMPLIED, OR STATUTORY WARRANTIES, INCLUDING, WITHOUT LIMITATION, ANY WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AVAILABILITY, ERROR-FREE OR UNINTERRUPTED OPERATION, AND ANY WARRANTIES ARISING FROM A COURSE OF DEALING, COURSE OF PERFORMANCE, OR USAGE OF TRADE. TO THE EXTENT THAT TECH LAB MAY NOT AS A MATTER OF APPLICABLE LAW DISCLAIM ANY IMPLIED WARRANTY, THE SCOPE AND DURATION OF SUCH WARRANTY WILL BE THE MINIMUM PERMITTED UNDER SUCH LAW. THE PRODUCTS AND SERVICES DO NOT CONSTITUTE BUSINESS OR LEGAL ADVICE. TECH LAB DOES NOT WARRANT THAT THE PRODUCTS AND SERVICES PROVIDED TO OR USED BY YOU HEREUNDER SHALL CAUSE YOU AND/OR YOUR PRODUCTS OR SERVICES TO BE IN COMPLIANCE WITH ANY APPLICABLE LAWS, REGULATIONS, OR SELF-REGULATORY FRAMEWORKS, AND YOU ARE SOLELY RESPONSIBLE FOR COMPLIANCE WITH THE SAME, INCLUDING, BUT NOT LIMITED TO, DATA PROTECTION LAWS, SUCH AS THE PERSONAL INFORMATION PROTECTION AND ELECTRONIC DOCUMENTS ACT (CANADA), THE DATA PROTECTION DIRECTIVE (EU), THE E-PRIVACY DIRECTIVE (EU), THE GENERAL DATA PROTECTION REGULATION (EU), AND THE E-PRIVACY REGULATION (EU) AS AND WHEN THEY BECOME EFFECTIVE.

# Glossary {#glossary}

| Term | Description |
| :---- | :---- |
| *Agent* | A software service encapsulated in a container that performs a specific, autonomous function (e.g., fraud detection, deal curation) within the bidstream. Agents are designed to operate with explicit intents and within orchestrator-defined constraints. |
| *Agent Manifest* | A JSON structure embedded as metadata in a container image that defines an agent’s capabilities, resource requirements, intents, and dependencies. Used by orchestrators to deploy and manage containers safely. |
| *Agent Orchestrator* | The control layer within a host platform that manages the deployment, execution, and coordination of agent containers participating in the bidstream. |
| *Agent to Agent (A2A)* | Direct communication or coordination between two autonomous agents, enabling them to exchange data or decisions without requiring mediation by a central orchestrator. |
| *Agentic System* | A distributed architecture where autonomous agents make decisions independently toward shared goals, such as optimizing auction outcomes, while maintaining defined constraints and interoperability. |
| *Autonomous Agentic* | Refers to systems or agents that make independent, AI-driven decisions and actions without direct orchestration, operating based on learned models or contextual goals. |
| *Bidstream* | The flow of bid requests and responses exchanged between programmatic advertising entities (SSPs, DSPs, exchanges). In ARTF, agents can modify bidstream data under orchestrator supervision. |
| *Bidstream Mutation*  | A proposed change (addition, deletion, or update) to data within the bidstream, expressed via an OpenRTB Patch. Each mutation is atomic and tied to a specific intent. |
| *Container* | A standardized, lightweight, and portable execution environment that packages an agent’s code, dependencies, and configuration into a single image. Containers conform to OCI runtime standards (e.g., Docker, Kubernetes). |
| *Intent* | A declarative statement describing why an agent proposes a mutation (e.g., “adjustBid,” “activateSegments,” “expireDeals”). Intents guide orchestrators’ decision-making about whether to accept or reject mutations. |
| *Manifest (Container Manifest)*  | The metadata file (usually container.json) defining container resources, intents, dependencies, and health checks, enabling orchestrators to deploy the container consistently. |
| *MCP (Model Context Protocol)*  | A protocol for structured model-to-agent communication using JSON-RPC. Complements gRPC and supports autonomic (AI-driven) agentic interactions, allowing models to orchestrate services directly. |
| *Mutation* | A single, atomic change proposed to a bid request or response. Includes fields such as intent, op (add/replace/remove), and path (semantic reference within OpenRTB payload). |
| *Orchestrator (Orchestrating Entity)* | The host platform (e.g., SSP, DSP, exchange) that manages the execution of agent containers, controls access, validates mutations, and decides whether to apply proposed changes. |
| *Patch* | A structured set of mutations proposed by an agent to modify an OpenRTB request or response. Each patch is atomic, traceable, and subject to orchestrator approval. |
| *Sidecar*  | A secondary container that runs alongside a primary application container, providing auxiliary functions (e.g., monitoring, telemetry, or mediation). |
| *Structural Agentic* | Agentic systems where interactions are between structured services (service-to-service orchestration), rather than direct AI model control. |
| *Telemetry*  | Monitoring data (metrics, traces, logs) emitted by containers to track performance, security, and decision impact across orchestrated systems. |
| *User Cohorts / Audience Segments* | Groupings of users with shared characteristics or behaviors that agents can activate or modify as part of audience segmentation use cases. |

# 

# Table of Contents {#table-of-contents}

[**About this document	2**](#about-this-document)

[**Glossary	5**](#glossary)

[**Table of Contents	6**](#table-of-contents)

[**Introduction	7**](#introduction)

[What is an agentic system?	7](#what-is-an-agentic-system?)

[Requirements	7](#requirements)

[What is a Container?	8](#what-is-a-container?)

[Why Containers?	8](#why-containers?)

[Integration into the Container Ecosystem	9](#integration-into-the-container-ecosystem)

[Why OpenRTB Patch?	9](#why-openrtb-patch?)

[Path Clarification	11](#path-clarification)

[Why gRPC and MCP?	11](#why-grpc-and-mcp?)

[**Agent Manifest	12**](#agent-manifest)

[**Technical Implementation	13**](#technical-implementation)

[Infrastructure Considerations	13](#infrastructure-considerations)

[Manifest Requirements	13](#manifest-requirements)

[Bidstream Integration	14](#bidstream-integration)

[Service Naming	15](#service-naming)

[API Design	15](#api-design)

[Example Extension Point	15](#example-extension-point)

[Request Message	16](#heading=h.cavzx66mi77z)

[AgenticRTB Specific Requirements	16](#agenticrtb-specific-requirements)

[Extension Response	17](#extension-response)

[Example \- Audience Segments \- Activating Cohorts	18](#example---audience-segments---activating-cohorts)

[Example \- Complex Orchestration	19](#example---complex-orchestration)

[Manifest \- Container.json	20](#manifest---container.json)

# Introduction {#introduction}

## What is an agentic system? {#what-is-an-agentic-system?}

The digital advertising ecosystem is built on a distributed system that allows for multi-party participation across a variety of business activities. This distributed system has common goals, but individual systems that are called by other partners to accomplish these goals. Each of these distributed systems also work to accomplish their own goals to accomplish their own mandates and make their own decisions to accomplish both the global goal (ie, a specific auction) as well as their own goals and mandates.

A **structurally agentic system** is a distributed architecture where autonomous components operate with explicit mandates, decision-making authority, and structured delegation patterns to accomplish domain-specific goals within defined constraints. Supply-Side Platforms (SSPs) and Demand-Side Platforms (DSPs) function as independent agents that orchestrate real-time auctions by delegating specialized tasks to other agents through standardized protocols such as OpenRTB. Each agent maintains internal state, optimizes toward distinct objectives (e.g., maximizing publisher yield vs. advertiser ROI), and makes autonomous decisions within strict latency budgets while the system's overall behavior emerges from the interaction of these specialized components rather than centralized control. This architectural pattern enables composability, fault isolation, independent optimization, and the ability to integrate new capabilities without modifying core auction logic, making it fundamentally different from monolithic systems where a single component possesses global knowledge and decision-making authority.

This standard introduces the ability to add new agents into existing orchestrations by introducing two new constructs \- containers and OpenRTB patch. The focus of this document is to enable Agentic extensibility for the RTB process, but both the container and the OpenRTB patch are a generic agentic approach, and can be used for non real-time bidding use cases as well.

## Requirements {#requirements}

For agentic extensibility to be adopted widely in the OpenRTB ecosystem, implementers need to adhere to a basic set of principles. These principles are:

1) **Agents must be able to participate in the core bidstream.** The bidstream may have many different points of integration with containers. This standard is focused on entities that are transacting via the bidstream in real-time, and is not a general-purpose standard for containerized services at the edge.   
2) **Agents must accomplish a specific goal.** Each agent must declare specific intents for itself, and any changes that it wants to make to auctions moving through the bidstream. The orchestrating entity may then choose to accept the change or not to facilitate multi-party participation.    
3) **Agents must be composable and deployable in standard enterprise infrastructure.** Agents must be structured as Containers and must adhere to the standard OCI runtime and image specifications (i.e., Docker Containers) so they are manageable via distributed cluster orchestration systems such as Kubernetes, Docker Compose / Swarm or cloud based systems like Amazon’s Elastic Containers service.   
4) **Agents must be performant and efficient**. Agents must communicate via a high performance protocol such as gRPC or in some cases MCP. Agents should be written in an efficient language and leverage an efficient ecosystem (for example, Rust, Go, or Java) or a highly optimized interpreted language. Agents should not consume resources they do not need.  Orchestrating entities must provide guidance on expected response times to container providers.  
5) **Agents must adhere to the policy of least-privilege and least-data. Agents will not have unfettered access to the outside world**, and must leverage appropriate services provided by the orchestrating entity. Orchestrating entities should implement appropriate measures to ensure containers adhere to this requirement.

## What is a Container? {#what-is-a-container?}

A container is a standard technology, originally popularized by Docker that provides an image-based, versioned, lightweight, and portable execution environment that encapsulates an application and all its dependencies—code, runtime, system tools, libraries, and settings—into a single package. This approach ensures that the containerized application behaves consistently across different computing environments, whether running locally, in a private data center, or in the cloud.

Containers provide a form of operating system-level virtualization where each container runs in isolation with its own runtime image, while sharing the host OS kernel. They are built using Open Container Initiative (OCI) standards, which define image formats and runtime specifications, making containers interoperable across various orchestration platforms such as Kubernetes, Docker Compose, or cloud-based services like Amazon Elastic Container Service (ECS).

In the context of OpenRTB and programmatic advertising, containers act as execution units for business logic that must interact with the bidstream in real-time. These execution units can host logic for bid modification, audience activation, deal curation, and other bidstream processes, while offering flexibility, modularity and simplified deployment for participants without requiring the container or the orchestrator to provide bespoke core platform infrastructure to each partner.

### Why Containers? {#why-containers?}

Containers enable a portable (capable of running in many environments), composable (can be combined with other containers or systems to provide value), dynamically scalable (capacity can be ramped up and down as demand does), and protected mode (manage ingress and egress, protect IP and sensitive data allowing execution of decisioning logic across diverse programmatic environments). Container’s standardized packaging allows partners to deploy once and operate across multiple orchestrating entities and cloud infrastructures without retooling, while composability ensures that logic from buyers, sellers, identity providers, and measurement vendors can be integrated into a cohesive, real-time auctioning workflow. Critically, containers encapsulate business logic in a self-contained image isolating proprietary IP while still allowing participation in bidstreams without requiring independent scaling. 

### Integration into the Container Ecosystem {#integration-into-the-container-ecosystem}

To run a container from a partner company in an orchestrator's infrastructure, both parties must address specific requirements and restrictions:

* The agent provider must ensure their application runs properly as a non-root user  
* The orchestrator will never run the container as a root user  
* The container must implement a [Kubernetes compatible health and readiness probes](https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/) (HTTP endpoints, TCP checks, or exec commands)  
* Containers must follow the principle of least privilege and remove all unnecessary capabilities.  
* The container must be built to handle graceful shutdowns, respect security contexts, and must not require privileged access or host network/PID namespaces.   
* The container must run without external network access: all network ingress and egress is prohibited with the exception of service communications with the orchestrating entity. Agent providers and orchestrating entities may negotiate specific network access policies to facilitate efficient operation of containers in the orchestrator’s environment.

On the orchestrating side, the orchestrator needs to implement appropriate RBAC policies that grant containers only the minimum required resource permissions. Network policies must be configured to control both ingress and egress traffic and explicitly define which services containers can communicate with and which ports are accessible. Depending on the required security level of network traffic control, the orchestrator may want to allow agent providers alternatives such as volume mappings with external data synchronization facilities to enable container data import and export. For secrets management, the orchestrator must securely inject any required credentials, API keys, or certificates using Kubernetes Secrets or similar methods, such as mounted secrets filesystem or secure environment variables) and will ensure they're properly scoped and rotated. Additionally, the orchestrator is expected to implement resource quotas and limits, use Pod Security Policies or Pod Security Standards or similar methods to enforce security constraints, and support monitoring and logging to track the container's behavior within their infrastructure.

### Why OpenRTB Patch? {#why-openrtb-patch?}

OpenRTB itself is a multi-party structurally agentic system. Complex interactions between many different parties are built into the OpenRTB and AdCOM domain models which describe the parameters of the real time bidding process. Introducing new agents into the existing system bypassing the full OpenRTB payload to those agents and the agent then returning the full, updated OpenRTB payload has a number of drawbacks:

* Inefficiency of passing the full OpenRTB payload to every entity introduces considerable serialization and deserialization costs.   
* Using a sequential approach for agentic calls has unacceptable performance implications  
* Running multiple agentic calls with the same payload that all return a modified payload makes it extraordinarily difficult and expensive to understand the changes to the payload that each individual agent makes.   
* The full OpenRTB payload exposes all information, including information that may not be germane to the agent's purpose or may contain competitive or restricted information.   
* Data may need to be scrubbed for data privacy purposes as well. Reconstructing the payload would be expensive.  
* Some information needed by the agent may not be part of the OpenRTB standard.   
* It may be very difficult for orchestrating systems to understand *why* some changes are made.

To address these changes, This standard envisions a protocol that meets the following requirements:

1) Each Agent should only get the minimum data allowed by the orchestrator and requested by the agent provider, following the least-data principle. In particular, private and competitive signals must be removed prior to agentic involvement.   
2) Agentic Changes to the OpenRTB payload must be isolated to ensure that each change can be evaluated or rejected independently.  
3) Each agent should be decoupled from the processing pipeline implementation. An agent implementation should never know internal specifics such as call sites or internal structures of orchestration systems.   
4) Agentic Changes to the OpenRTB payload must declare their intentions to ensure that orchestrators understand why proposed changes are made. 

To meet these requirements this standard extends the model by introducing a patching mechanism, the *OpenRTB Patch Protocol*, which provides a standardized protocol for requesting changes to an OpenRTB bid request or bid response. When processing bid requests and bid responses, the orchestrator will identify the containers that should be included in its processing and will issue a request to each. Participant responses may request in-flight modifications via patch objects that include desired mutations which the orchestrator may accept or reject at their discretion. The OpenRTB Patch Protocol also introduces a concept of *intents* which are declarative attributes that indicate the purpose of a given mutation which enables orchestrators to better determine when and how containers should be invoked . 

![][image2]

### **Path Clarification** {#path-clarification}

Sometimes agents might return mutations that need to be inserted at a later point, and may specify a particular id (for example, a particular deal-id). In some cases mutations might be inserted in a different location or sequence than the agent anticipates. To maintain this flexibility on an intention by intention basis \- mutation paths may use semantic references derived from openRTB concepts. Rather than following explicit paths to the overall RTB structure (which the agent may not know), paths identify business level entities rather than specific JSON locations. This allows the spec to express mutations in terms of auction semantics rather than data layout.

### **Why gRPC and MCP?** {#why-grpc-and-mcp?}

This spec for agents requires that containers support Remote Procedure Calls for communication between orchestrators and agents. GRPC with protobuf is mandated for all interfaces in this document, while MCP may be used when appropriate. While OpenRTB recently added support for GRPC at a native level, most OpenRTB payloads are still REST and JSON. The reasons for supporting GRPC instead of REST are:

1) Performance needs to be as fast as possible for these interfaces. GRPC uses protobuf serialization, which is considerably more space and time efficient than JSON.  
2) GRPC is easy to validate and has built-in rejection of invalid payloads.   
3) Remote Procedure Calls are a better model for agent-based communication than state transitions. 

MCP may also be used (and will likely be required in the future). In particular MCP version 2025-06-18 is the first version that is both performant enough and secure enough to handle significant traffic. This is primarily due to the support of streamable HTTP as well as OAuth authentication. MCP also uses JSON-RPC which makes it a natural successor to REST based interfaces. In addition, while GRPC is ideal for service to service agentic orchestration, MCP can be used not only for service to service orchestration, but also model to agent orchestration, enabling autonomic agentic flows as well. 

In both cases, the same set of interfaces are exposed \- in GRPC via rpc definitions in protobuf, or via tool definitions in MCP. 

# Agent Manifest {#agent-manifest}

An agent manifest is a standard field in JSON added to the container image via the image metadata expressed as a label which defines how a container is used and managed by an orchestrator. The manifest is specified as a key-value pair with the key being “agent-manifest” and the value being a JSON structure. 

The manifest provides business requirements for the container such as:

* The name of the agent  
* The vendor for the agent  
* The owner (email address) of the agent

It also includes a number of business metadata fields:

* The intent(s) supported by this agent. 

It also includes information about the runtime configuration for the container:

* The minimum CPU and memory resources it requires,   
* Any dependencies on other services (by name) that the container needs to communicate with. 

# Technical Implementation {#technical-implementation}

## Infrastructure Considerations {#infrastructure-considerations}

Composable and deployable mean that Agents need to be able to be deployed in a wide variety of infrastructure in a manner that is consistent with the base infrastructure. In practice this means that they must be structured as OCI containers and must be deployable into a distributed container ecosystem (typically Kubernetes)  
The following best practices for Docker images must be followed by both the agent provider and orchestrator for any agent acting on bidstream data:

* **Mandatory health checks** should be implemented as both liveness and readiness probes to ensure robust self-healing capabilities, allowing infrastructure to operate autonomously and maintain continuous service availability even during partial failures.  
* **Image signing and verification** must be enforced across all deployments to establish a chain of trust from development to production, preventing unauthorized or compromised artifacts from entering the bidstream ecosystem.  
* **Images must be built by a managed and auditable CI/CD pipeline.** Pipelines should automate the entire release process with integrated security scanning, compliance checks, and deployment validation to maintain velocity while ensuring quality and consistency.  
* **Standard logging and monitoring** must be integrated into all deployments to provide real-time visibility into system health, enabling proactive identification of performance bottlenecks and security anomalies before they impact service delivery. Containers must support Open Telemetry endpoints for metrics and support Open Tracing for distributed trace and span collections.

The following are a few of the best practices for docker images which must be followed for any orchestrator:

* **Network policies and segmentation** must be implemented to control traffic flows between services, between containers and between privacy sensitive services to protect them from unauthorized workloads. Systems should be configured to provide defense-in-depth isolation that prevents lateral movement and limits the blast radius of security breaches within your infrastructure.  
* **Resource quotas and limits** must  be defined at both namespace and container levels to guarantee predictable performance, prevent resource starvation, and optimize utilization across your infrastructure.

## Manifest Requirements {#manifest-requirements}

Agents are expected to provide a manifest as described above that expresses their needs:

1) Minimum CPU/Memory requirements  
2) The intents they will invoke  
3) Other services that they expect to be available (by name)

If the orchestrating entity cannot support the requirements and intents of a container, it should not be started. 

## Bidstream Integration {#bidstream-integration}

Agents are implemented as containers that are essentially black boxes that allow service providers to execute core business logic. Each container “lives” in a detached and isolated compute environment once deployed. For these containers to work, they must integrate into the bid stream using predefined intentions. Because participants in the AdTech ecosystem generally have unique mixes of infrastructure, software, and practices. What extension points are supported and how they integrate with agents will likely be custom for each participant. Since it’s not feasible or desirable to have everyone run the same software, instead, a standard protocol is defined that supports generic interactions with extension points. OpenRTB provides robust support for data communication within the bidstream, with appropriate abstractions for vendor-specific extensions, but lacks standardized support for incorporating changes like those containers may want to propagate back to the bid stream as part of a decisioning cycle. To address that shortfall, this standard introduces “OpenRTB Patch,” a protocol for expressing desired changes back into OpenRTB requests and responses. 

OpenRTB Patch is designed to adhere to the following principles:

* Containers provide desired mutations to the orchestrator as “patches” which describe a delta from the provided request or response and may include some combination of additions, changes, and deletions to specified portions of the request or response. Whether or not a patch is applied is left to the discretion of the orchestrator, which will decide to accept or reject mutations in accordance with its business requirements. Whether and how orchestrators will communicate choices to accept or reject patches to agent providers is left to the parties.  
* Each patch is atomic. A mutation must be accepted in whole or rejected in whole. Multiple patches may be independently accepted or rejected \- there are no transactions or any ordering guarantees across mutations.    
* Mutations are semantically meaningful: an OpenRTB Patch specifies not only what change is desired, but also why it is requested \- this is specified as the intent.   
* OpenRTB patch propagates necessary privacy and signal information into containers. It is the container's responsibility to honor all of the specified behavior in the RTB request.  
* OpenRTB patch follows the semantics of the OpenRTB standard, including use of “ext” objects for any nonstandard, extended signaling orchestrators that may wish to make available to agent providers.   
* OpenRTB patch facilitates auditability within an orchestrator's environment by logging requested and applied patches.  
* Orchestrators should provide standard reports and metrics to judge the acceptance rate of mutations and rejection reasons. 

## Service Naming {#service-naming}

As mentioned above, containers can provide specific dependencies in the name of service that they have a dependency on, Since this will often map to proprietary services that orchestrators provide, or names of other containers, the specification is simply that the service dependency be a string to signify the dependency for mapping. 

# API Design {#api-design}

RPC definitions for both gRPC and for MCP require a bit more structure than a traditional REST/JSON service. For purposes of this document, a few key patterns are defined, but the authoritative definition for this is the gRPC definition found in the IABTechLab’s Github Project.  
These definitions take the form of Protobuf specifications. In particular, the services and messages must be articulated in .proto files. For MCP use cases, the specification is the same as GRPC, with the obvious exception of a tool definition rather than RPC endpoints for each extension point. 

For the purpose of the comment period, there is an initial protobuf definition. These are subject to change. 

## Example Extension Point {#example-extension-point}

syntax \= "proto2";

package com.iabtechlab.bidstream.mutation.v1;

import "com/iabtechlab/openrtb/v2.6/openrtb.proto";

// service definition  
service RTBExtensionPoint {  
  // GetMutations returns RTBResponse containing mutations to be applied at the predetermined auction lifecycle event    
  rpc GetMutations (RTBRequest) returns (RTBResponse);  
}

message RTBRequest {  
  // ENUM as per Programmatic Auction Definition IAB TL doc/spec  
  required Lifecycle lifecycle \= 1;

  required string id \= 2;  
  optional Extensions ext \= 3;

  required com.iabtechlab.openrtb.v2.BidRequest bid\_request \= 4;

  optional com.iabtechlab.openrtb.v2.BidResponse bid\_response \= 5;

  required int32 tmax \= 6;

}

This forms the backbone of the extension system. Each individual request exposes a single rpc and a single stream mechanism for a given extension point. Right now the spec does anticipate the need to support multiple called endpoints so that the same container could support multiple different service requests. 

### **AgenticRTB Specific Requirements** {#agenticrtb-specific-requirements}

| Field | Type | Description |
| ----- | ----- | ----- |
| id | string | Extension point Request ID |
| bidRequest.imp | imp message | Impression message (per OpenRTB, with exceptions) |
| bidRequest.site | site message | Site message (per OpenRTB) |
| bidRequest.app | app message | App message (per OpenRTB) |
| bidRequest.device | device message | Device message (per OpenRTB) |
| bidRequest.user | user message | User message (per OpenRTB) |
| bidRequest.regs | reg message | Regulation message (per OpenRTB) |
| bidRequest.source | source message | Source message (per OpenRTB) |
| bidRequest.tmax | integer | Maximum time in milliseconds the exchange allows for mutations to be received including internet latency to avoid timeout |

### Extension Response {#extension-response}

A mutation represents a change to an existing request —it modifies the system’s state by adding, removing, or updating records. Extension Provider responses take the form of these mutations, proposing adjustments to bid requests or responses. An example Extension Response gRPC is below. 

`message RTBResponse { // Or RequestPatch`  
  `string id = 1;`  
  `repeated mutation mutations = 2;`  
  `MetaData metadata = 3;`  
`}`

`message Mutation {`  
  `string intent = 1;`  
  `string op = 2;`  
  `string path = 3;`

  `// The structure of value depends on the specified intent.`  
  `// Reserve 100+ for intent-specific payloads`  
 `// The structure of value depends on the specified intent.`  
  `// Reserve 100+ for intent-specific payloads`  
  `oneof value {`  
    `// List of string Identifiers`  
    `IDsPayload ids = 100;`

    `// Adjust properties of a specific deal`  
    `AdjustDealPayload adjust_deal = 101;`

    `// Adjust the bid price`  
    `AdjustBidPayload adjust_bid = 102;`

    `// Metrics or telemetry data`  
    `AddMetricsPayload add_metrics = 103;`  
  `}`

}

`message MetaData {`  
  `string api_version = 1;`  
  `string model_version = 2;`  
`}`

## Example \- Audience Segments \- Activating Cohorts {#example---audience-segments---activating-cohorts}

Below is an example of a response return. Since gRPC is a binary protocol, JSON format is used to express concepts rather than serialization formats. 

`{`  
    `"intent": "activateSegments",`  
    `"op": "add",`  
    `"path": "/user/data/segment",`  
    `"value": {`  
        `"IDsPayload": [“18-35-age-segment","soccer-watchers"]`  
    `}`  
`}`

## Example \- Complex Orchestration {#example---complex-orchestration}

Endpoints may return multiple mutations. If so, each mutation is evaluated separately. Note that there is no support for transactions across multiple mutations. Any mutation here may be accepted or rejected independently of other mutations. 

`[{`  
    `"intent": "expireDeals",`  
    `"op": "remove"`  
    `"path": "/imp/1",`  
    `"value": {`  
        `"IDsPayload": [“deal100","deal200"]`  
    `}`  
`},{`  
    `"intent": "activateDeals",`  
    `"op": "add",`  
    `"path": "/imp/1",`  
    `"value": {`  
        `"IDsPayload": [“deal300","deal201"]`  
    `}`  
`},{`  
    `"intent": "adjustDeals",`  
    `"op": "replace"`  
    `"path": "/imp/2/deals/400",`  
    `"value": {`  
        `“AdjustDealPayload”: {`  
            `"bidfloor": 5.00,`  
            `"wadomain": ["adomain.com"],`  
        `}`  
     `},`  
`},{`  
    `"intent": "adjustDeals",`  
    `"op": "replace",`  
    `"path": "/imp/1/deals/500",`  
    `"value": {`  
        `“AdjustDealPayload”: {`  
            `"bidfloor": 8.00`  
        `}`  
    `},`  
`}]`

## Manifest \- Container.json {#manifest---container.json}

This is an example of a container JSON. Note that this is associated with the metadata for the docker image for transmission. 

{  
  "name": "openrtb-container-suite",  
  "version": "1.0.0",  
  "description": "Example container manifest for OpenRTB and Digital Advertising use cases",  
  "image": {  
    "repository": "registry.example.com/rtb/auction-container",  
    "tag": "v1.0.5",  
    "digest": "sha256:abc123def4567890"  
  },  
  "resources": {  
    "cpu": "500m",  
    "memory": "256Mi"  
  },  
  "intents": \[  
    {  
      "name": "bidResponseGeneration",  
      "description": "Generate bid responses in real time based on request data"  
    },  
    {  
      "name": "bidValuation",  
      "description": "Evaluate incoming bid requests and determine bid amount"  
    },  
    {  
      "name": "bidRequestModification",  
      "description": "Propose mutations to a bid request prior to auction execution"  
    },  
    {  
      "name": "auctionOrchestration",  
      "description": "Route or prioritize bid requests across multiple buyers"  
    },  
    {  
      "name": "metadataEnhancement",  
      "description": "Insert or modify auction metadata such as fraud or viewability signals"  
    },  
    {  
      "name": "dynamicDealCuration",  
      "description": "Curate deals in real time, optimize margins or enforce dynamic inclusion/exclusion lists"  
    },  
    {  
      "name": "audienceSegmentation",  
      "description": "Activate or enrich user cohorts and audience segments"  
    }  
  \],  
  "dependencies": {  
    "fraudDetectionService": {  
      "service": "fraud-svc",  
      "ENV\_VARIABLE": FRAUD\_URL  
    },  
    "audienceService": {  
      "host": "audience-svc",  
      "port": 9010  
    }  
  },  
  "health": {  
    "livenessProbe": {  
      "httpGet": {  
        "path": "/health/live",  
        "port": 8080  
      },  
      "initialDelaySeconds": 10,  
      "periodSeconds": 5  
    },  
    "readinessProbe": {  
      "httpGet": {  
        "path": "/health/ready",  
        "port": 8080  
      },  
      "initialDelaySeconds": 5,  
      "periodSeconds": 5  
    }  
  },  
  "security": {  
    "runAsNonRoot": true,  
    "dropCapabilities": \["NET\_ADMIN", "SYS\_PTRACE"\],  
    "networkPolicies": {  
      "ingress": \["fraud-svc", "audience-svc"\],  
      "egress": \["logging-svc"\]  
    }  
  },  
  "maintainers": \[  
    {  
      "name": "IAB Tech Lab Example Team",  
      "email": "support@iabtechlab.com"  
    }  
  \]  
}

[image1]: <data:image/png;base64,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>

[image2]: <data:image/png;base64,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>
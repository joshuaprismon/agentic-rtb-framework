use std::pin::Pin;
use std::{env, str};
use futures::Stream;
use tokio_stream::StreamExt;
use tonic::{transport::Server, Request, Response, Status};
use hyper::{Method, StatusCode, Server as HyperServer};
use hyper::service::{make_service_fn, service_fn};
use std::convert::Infallible;
use std::net::SocketAddr;

pub mod com {
    pub mod iabtechlab {
        pub mod bidstream {
            pub mod mutation {
                pub mod v1 {
                    include!("com.iabtechlab.bidstream.mutation.v1.rs");
                }
            }
        }
        pub mod openrtb {
            pub mod v2_6 {
                include!("com.iabtechlab.openrtb.v2_6.rs");
            }
        }
    }
}

use com::iabtechlab::bidstream::mutation::v1::*;
use crate::rtb_extension_point_server::RtbExtensionPointServer;

use crate::mutation::Value::Ids;
use crate::mutation::Value::AdjustBid;

const API_VERSION: &str = "1.0.0";
const MODEL_VERSION: &str = "v0.10.0";

// Evaluate the RtbRequest and return a RtbResponse
async fn evaluate(req: RtbRequest) -> RtbResponse {
    // For demonstration purposes, we will create a static response
    // In a real-world scenario, you would implement logic to evaluate the request
    // and determine the appropriate mutations based on the request data.
    let metadata = Metadata {
        api_version: Some(API_VERSION.to_string()),
        model_version: Some(MODEL_VERSION.to_string())
    };

    match req.id.as_str() {
        "auction-123" => RtbResponse {
            id: req.id,
            mutations: vec![
                Mutation {
                    intent: Intent::ActivateSegments.into(),    // Activate segment(s) for the user
                    op: Operation::Add.into(),  // Add the segment(s) to the user
                    path: "/user/data/segment".to_string(), // Path to the user data segment in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the segment IDs to be added
                        id: vec![
                            "seg-sports".to_string(), 
                            "demo-25-35".to_string(),
                            "gender-male".to_string()
                            ],
                    })),
                },
                Mutation {
                    intent: Intent::ActivateDeals.into(),   // Activate deal(s) to the impression
                    op: Operation::Add.into(),  // Add the deal(s) to the impression
                    path: "/imp/imp-1".to_string(), // Path to the impression in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the deal IDs to be added
                        id: vec!["display-deal-001".to_string()],
                    }))
                }
            ],
            metadata: Some(metadata)
        },
        "auction-456" => RtbResponse {
            id: req.id,
            mutations: vec![
                Mutation {
                    intent: Intent::ActivateSegments.into(),    // Activate segment(s) for the user
                    op: Operation::Add.into(),  // Add the segment(s) to the user
                    path: "/user/data/segment".to_string(), // Path to the user data segment in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the segment IDs to be added
                        id: vec!["demo-35-44".to_string()]
                    })),
                },
                Mutation {
                    intent: Intent::ActivateDeals.into(),   // Activate deal(s) to the impression
                    op: Operation::Add.into(),  // Add the deal(s) to the impression
                    path: "/imp/imp-1".to_string(), // Path to the impression in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the deal IDs to be added
                        id: vec![
                            "premium-deal-001".to_string(), 
                            "video-deal-001".to_string()
                            ]
                    }))
                }
            ],
            metadata: Some(metadata)
        },
        "auction-789" => RtbResponse {
            id: req.id,
            mutations: vec![
                Mutation {
                    intent: Intent::ActivateSegments.into(),    // Activate segment(s) for the user
                    op: Operation::Add.into(),  // Add the segment(s) to the user
                    path: "/user/data/segment".to_string(), // Path to the user data segment in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the segment IDs to be added
                        id: vec!["demo-45-plus".to_string()],
                    })),
                },
                Mutation {
                    intent: Intent::ActivateDeals.into(),   // Activate deal(s) to the impression
                    op: Operation::Add.into(),  // Add the deal(s) to the impression
                    path: "/imp/imp-1".to_string(), // Path to the impression in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the deal IDs to be added
                        id: vec!["display-deal-001".to_string()],
                    }))
                },
                Mutation {
                    intent: Intent::BidShade.into(),   // Activate deal(s) to the impression
                    op: Operation::Replace.into(),  // Add the deal(s) to the impression
                    path: "/seatbid/dsp-001/bid/bid-abc".to_string(), // Path to the impression in ORTB object
                    value: Some(AdjustBid(AdjustBidPayload {
                        price: Some(4.675) 
                    }))
                }
            ],
            metadata: Some(metadata)
        },
        "auction-multi-123" => RtbResponse {
            id: req.id,
            mutations: vec![
                Mutation {
                    intent: Intent::ActivateSegments.into(),    // Activate segment(s) for the user
                    op: Operation::Add.into(),  // Add the segment(s) to the user
                    path: "/user/data/segment".to_string(), // Path to the user data segment in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the segment IDs to be added
                        id: vec![
                            "demo-35-44".to_string(), 
                            "gender-female".to_string()
                            ],
                    })),
                },
                Mutation {
                    intent: Intent::ActivateDeals.into(),   // Activate deal(s) to the impression
                    op: Operation::Add.into(),  // Add the deal(s) to the impression
                    path: "/imp/imp-header".to_string(), // Path to the impression in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the deal IDs to be added
                        id: vec!["display-deal-001".to_string()],
                    })),
                },
                Mutation {
                    intent: Intent::ActivateDeals.into(),   // Activate deal(s) to the impression
                    op: Operation::Add.into(),  // Add the deal(s) to the impression
                    path: "/imp/imp-sidebar".to_string(), // Path to the impression in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the deal IDs to be added
                        id: vec!["display-deal-001".to_string()],
                    })),
                },
                Mutation {
                    intent: Intent::ActivateDeals.into(),   // Activate deal(s) to the impression
                    op: Operation::Add.into(),  // Add the deal(s) to the impression
                    path: "/imp/imp-footer".to_string(), // Path to the impression in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the deal IDs to be added
                        id: vec!["display-deal-001".to_string()],
                    })),
                }
            ],
            metadata: Some(metadata)
        },
        "app-123" => RtbResponse {
            id: req.id,
            mutations: vec![
                Mutation {
                    intent: Intent::ActivateSegments.into(),    // Activate segment(s) for the user
                    op: Operation::Add.into(),  // Add the segment(s) to the user
                    path: "/user/data/segment".to_string(), // Path to the user data segment in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the segment IDs to be added
                        id: vec!["demo-18-24".to_string()],
                    })),
                },
                Mutation {
                    intent: Intent::ActivateDeals.into(),   // Activate deal(s) to the impression
                    op: Operation::Add.into(),  // Add the deal(s) to the impression
                    path: "/imp/imp-1".to_string(), // Path to the impression in ORTB object
                    value: Some(Ids(IDsPayload {    // Payload containing the deal IDs to be added
                        id: vec!["native-deal-001".to_string()],
                    })),
                }
            ],
            metadata: Some(metadata)
        },
        _ => RtbResponse {
            id: req.id,
            mutations: vec![],
            metadata: Some(metadata),
        }
    }
}

async fn handle_rest(req: hyper::Request<hyper::body::Body>) -> Result<hyper::Response<hyper::body::Body>, Infallible> {
    match (req.method(), req.uri().path()) {
        (&Method::GET, "/health/live") => {
            Ok(hyper::Response::new(hyper::body::Body::from("OK")))
        },
        (&Method::GET, "/health/ready") => {
            Ok(hyper::Response::new(hyper::body::Body::from("OK")))
        },
        _ => {
            Ok(hyper::Response::builder()
                .status(StatusCode::NOT_FOUND)
                .body(hyper::body::Body::from("Not Found"))
                .unwrap())
        }
    }
}

#[derive(Default)]
pub struct RtbExtensionPointService {}

#[tonic::async_trait]
impl rtb_extension_point_server::RtbExtensionPoint for RtbExtensionPointService {
    async fn get_mutations (
        &self,
        request: Request<RtbRequest>,
    ) -> Result<Response<RtbResponse>, Status> {
        let response: RtbResponse = evaluate(request.into_inner()).await;
        Ok(Response::new(response))
    }

    type GetMutationStreamStream = Pin<Box<dyn Stream<Item = Result<RtbResponse, Status>> + Send + 'static>>;

    async fn get_mutation_stream(
        &self,
        request: Request<tonic::Streaming<RtbRequest>>,
    ) -> Result<Response<Self::GetMutationStreamStream>, Status> {
        let mut input_stream: tonic::Streaming<RtbRequest> = request.into_inner();
        let output_stream = async_stream::try_stream! {
            while let Some(request) = input_stream.next().await {
                let response = evaluate(request?).await;
                yield response;
            }
        };

        Ok(Response::new(Box::pin(output_stream) as Self::GetMutationStreamStream))
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_prost_build::compile_protos("proto/agenticrtbframework.proto").unwrap_or_else(|e| panic!("Failed to compile protos {:?}", e));

    // gRPC server environment variables
    let address = env::var("ARTF_GRPC_SERVER_ADDRESS").unwrap_or_else(|_| "0.0.0.0".to_string()).parse::<String>().unwrap();
    let grpc_port = env::var("ARTF_GRPC_SERVER_PORT").unwrap_or_else(|_| "50051".to_string()).parse::<u16>().unwrap();
    let http_port = env::var("ARTF_HTTP_SERVER_PORT").unwrap_or_else(|_| "8080".to_string()).parse::<u16>().unwrap();
    let max_server_connection: u16 = env::var("ARTF_MAX_CONNS").unwrap_or_else(|_| "256".to_string()).parse::<u16>().unwrap();
   
    let grpc_addr = format!("{}:{}", address, grpc_port).parse().unwrap();
    let agentic_rtb_framework_service = RtbExtensionPointServer::new(RtbExtensionPointService::default());

    println!("Agentic RTB Framework API Version: {}", API_VERSION);
    println!("Agentic RTB Framework Model Version: {}", MODEL_VERSION);
    println!("Setting gRPC Server Max connections: {}", max_server_connection);
    println!("Starting gRPC Server at: {}:{}", address, grpc_port);
    println!("Starting HTTP Server at: {}:{}", address, http_port);

    let grpc_server = tokio::spawn(async move{
        Server::builder()
            .concurrency_limit_per_connection(max_server_connection as usize)
            .add_service(agentic_rtb_framework_service)
            .serve(grpc_addr)
            .await

    });

    let rest_server = tokio::spawn(async move {
        let make_svc = make_service_fn(|_conn| async {
            Ok::<_, Infallible>(service_fn(handle_rest))
        });

        let rest_addr: SocketAddr = format!("{}:{}", address, http_port).parse().unwrap();

        HyperServer::bind(&rest_addr)
            .serve(make_svc)
            .await
    });

    // Wait for both servers (or handle their errors)
    tokio::select! {
        result = grpc_server => {
            match result {
                Ok(_) => println!("gRPC Server stopped successfully."),
                Err(e) => eprintln!("gRPC Server encountered an error: {:?}", e),
            }
        }
        result = rest_server => {
            match result {
                Ok(_) => println!("HTTP Server stopped successfully."),
                Err(e) => eprintln!("HTTP Server encountered an error: {:?}", e),
            }
        }
    }

    Ok(())
}

use std::pin::Pin;
use std::{env, str};
use futures::Stream;
use tokio_stream::StreamExt;
use tonic::{transport::Server, Request, Response, Status};

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

use com::iabtechlab::bidstream::mutation::v1::rtb_extension_point_server::{
    RtbExtensionPoint, RtbExtensionPointServer
};
use com::iabtechlab::bidstream::mutation::v1::{
    RtbRequest, RtbResponse, Mutation, Operation, Intent, Metadata
};


const VERSION: &'static str = "0.1.0";

// Evaluate the RTBRequest and return a RtbResponse
async fn evaluate(req: RtbRequest) -> RtbResponse {
    let result: RtbResponse = com::iabtechlab::bidstream::mutation::v1::RtbResponse {
        id: req.id,
        mutations: Vec::<Mutation>::new(),
        metadata: Some(Metadata::default())
    };

    result
}

#[derive(Default)]
pub struct RtbExtensionPointService {}

#[tonic::async_trait]
impl RtbExtensionPoint for RtbExtensionPointService {
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
    let address = env::var("ARTF_SERVER_ADDRESS").unwrap_or_else(|_| "0.0.0.0".to_string()).parse::<String>().unwrap();
    let port = env::var("ARTF_SERVER_PORT").unwrap_or_else(|_| "50051".to_string()).parse::<u16>().unwrap();
    let max_server_connection: u16 = env::var("ARTF_MAX_CONNS").unwrap_or_else(|_| "256".to_string()).parse::<u16>().unwrap();
   
    let addr = format!("{}:{}", address, port).parse().unwrap();
    let agentic_rtb_framework_service = RtbExtensionPointServer::new(RtbExtensionPointService::default());

    println!("Agentic RTB Framework gRPC Server Version: {}", VERSION);
    println!("Setting gRPC Server Max connections: {}", max_server_connection);
    println!("Starting gRPC Server at: {}:{}", address, port);

    Server::builder()
        .concurrency_limit_per_connection(max_server_connection as usize)
        .add_service(agentic_rtb_framework_service)
        .serve(addr)
        .await?;

    Ok(())
}

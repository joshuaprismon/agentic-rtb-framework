//! gRPC service implementation for the RTB extension point.

use futures::Stream;
use std::pin::Pin;
use tokio_stream::StreamExt;
use tonic::{Request, Response, Status};

use crate::bidder::evaluate;
use crate::proto::com::iabtechlab::bidstream::mutation::v1::{
    rtb_extension_point_server, RtbRequest, RtbResponse,
};

/// gRPC service that dispatches requests to the bidder evaluator.
#[derive(Default)]
pub struct RtbExtensionPointService {}

#[tonic::async_trait]
impl rtb_extension_point_server::RtbExtensionPoint for RtbExtensionPointService {
    async fn get_mutations(
        &self,
        request: Request<RtbRequest>,
    ) -> Result<Response<RtbResponse>, Status> {
        let response: RtbResponse = evaluate(request.into_inner()).await;
        Ok(Response::new(response))
    }

    type GetMutationStreamStream =
        Pin<Box<dyn Stream<Item = Result<RtbResponse, Status>> + Send + 'static>>;

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

        Ok(Response::new(
            Box::pin(output_stream) as Self::GetMutationStreamStream
        ))
    }
}

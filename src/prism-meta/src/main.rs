use std::io;

use async_once_cell::OnceCell;
use clap::Parser;
use envconfig::Envconfig;
use prism_proto::{
    meta_service_server::{MetaService, MetaServiceServer},
    GetTablePartitionsRequest, GetTablePartitionsResponse, GetTableSchemaRequest,
    GetTableSchemaResponse, TableColumn,
};
use sqlx::{postgres::PgPoolOptions, Pool, Postgres, Row};
use tonic::{transport::Server, Request, Response, Status};
use tracing::Level;
use tracing_subscriber::prelude::*;

#[derive(Envconfig, Debug)]
struct DbConfig {
    #[envconfig(from = "COCKROACHDB_USER")]
    pub cockroach_user: String,
    #[envconfig(from = "COCKROACHDB_PASSWORD")]
    pub cockroach_password: String,
    #[envconfig(from = "COCKROACHDB_URL")]
    pub cockroach_url: String,
    #[envconfig(from = "COCKROACHDB_DATABASE")]
    pub cockroach_database: String,
}

static POOL: OnceCell<Pool<Postgres>> = OnceCell::new();

pub async fn initialize() -> anyhow::Result<()> {
    let config = DbConfig::init_from_env()?;
    POOL.get_or_try_init(async {
        PgPoolOptions::new()
            .max_connections(5)
            .connect(&format!(
                "postgres://{}:{}@{}/{}?sslmode=verify-full",
                config.cockroach_user,
                config.cockroach_password,
                config.cockroach_url,
                config.cockroach_database,
            ))
            .await
    })
    .await?;

    Ok(())
}

#[derive(Envconfig, Parser, Debug)]
struct Args {
    #[arg(short, long, default_value_t = 8080)]
    port: u16,
}

pub struct Meta {}

#[tonic::async_trait]
impl MetaService for Meta {
    async fn get_table_schema(
        &self,
        req: Request<GetTableSchemaRequest>,
    ) -> Result<Response<GetTableSchemaResponse>, Status> {
        let db = POOL.get().ok_or_else(|| {
            tracing::error!("no database connection");
            Status::internal("internal error")
        })?;

        let inner = req.into_inner();
        let rows = sqlx::query(
            r#"
            SELECT column_name, type FROM meta.table_schemas
            WHERE table_name = $1
            "#,
        )
        .bind(&inner.table_name)
        .fetch_all(db)
        .await
        .map_err(|e| Status::internal(format!("database error: {:?}", e)))?;

        let mut columns: Vec<TableColumn> = vec![];
        for row in rows {
            let column_name: String = row.get(0);
            let column_type: i32 = row.get(1);
            columns.push(TableColumn {
                name: column_name,
                r#type: column_type,
            });
        }

        Ok(Response::new(GetTableSchemaResponse {
            table_name: inner.table_name,
            columns,
        }))
    }

    async fn get_table_partitions(
        &self,
        req: Request<GetTablePartitionsRequest>,
    ) -> Result<Response<GetTablePartitionsResponse>, Status> {
        let db = POOL.get().ok_or_else(|| {
            tracing::error!("no database connection");
            Status::internal("internal error")
        })?;

        let inner = req.into_inner();
        let rows = if let Some(range) = inner.time_range {
            //
            // a-------b         a <= c && b <= d
            //     c------d
            //
            //    a-------b      a >= c && b <= d
            // c----d
            // a------b          a <= c && b >= d
            //   c--d
            //
            //   a--b            a >= c && b <= d
            // c-------d
            sqlx::query(
                r#"
                SELECT partition_name FROM meta.table_partitions
                WHERE table_name = $1
                    AND (
                        start_time <= $2 AND end_time <= $3
                        OR start_time <= $2 AND end_time >= $3
                        OR start_time >= $2 AND end_time <= $3
                        OR start_time >= $2 AND end_time >= $3
                    )
                "#,
            )
            .bind(&inner.table_name)
            .bind(range.start_time)
            .bind(range.end_time)
            .fetch_all(db)
            .await
            .map_err(|e| Status::internal(format!("database error: {:?}", e)))?
        } else {
            sqlx::query(
                r#"
                SELECT partition_name FROM meta.table_partitions
                WHERE table_name = $1
            "#,
            )
            .bind(&inner.table_name)
            .fetch_all(db)
            .await
            .map_err(|e| Status::internal(format!("database error: {:?}", e)))?
        };

        let mut partitions: Vec<String> = vec![];
        for row in rows {
            partitions.push(row.get(0));
        }

        Ok(Response::new(GetTablePartitionsResponse {
            table_name: inner.table_name,
            partitions,
        }))
    }
}

#[tokio::main]
async fn main() {
    let layer = tracing_subscriber::fmt::layer()
        .with_writer(io::stderr)
        .and_then(
            tracing_subscriber::EnvFilter::from_default_env().add_directive(Level::INFO.into()),
        )
        .boxed();
    tracing_subscriber::registry().with(layer).init();

    if let Err(e) = initialize().await {
        tracing::error!(err = ?e, "error initializing database");
        std::process::exit(1);
    }

    let args = Args::parse();
    let addr = format!("127.0.0.1:{}", args.port).parse().unwrap();
    let service = Meta {};
    tracing::info!(port = args.port, "starting server");
    Server::builder()
        .add_service(MetaServiceServer::new(service))
        .serve(addr)
        .await
        .unwrap();
}

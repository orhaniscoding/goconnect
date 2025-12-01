fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Compile protobuf definitions
    tonic_build::configure()
        .build_server(false) // We only need the client
        .compile_protos(
            &["../../core/proto/daemon.proto"],
            &["../../core/proto"],
        )?;

    tauri_build::build();
    Ok(())
}

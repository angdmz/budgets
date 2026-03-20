from logging.config import fileConfig
import os

from sqlalchemy import engine_from_config, pool, URL
from alembic import context

from entities import metadata
from secrets import get_secrets_provider, SecretNotFoundError

config = context.config

if config.config_file_name is not None:
    fileConfig(config.config_file_name)

target_metadata = metadata


def get_env_or_default(key: str, default: str) -> str:
    return os.environ.get(key, default)


def load_database_config():
    """Load database configuration using secrets provider."""
    provider = get_secrets_provider()
    
    # Get secrets
    try:
        db_password = provider.get_secret("db_password")
    except SecretNotFoundError:
        db_password = get_env_or_default("DB_PASSWORD", "postgres")
    
    try:
        encryption_key = provider.get_secret("encryption_key")
    except SecretNotFoundError:
        encryption_key = get_env_or_default("ENCRYPTION_KEY", "")
    
    return {
        "url_prefix": get_env_or_default("DB_URL_PREFIX", "postgresql"),
        "hostname": get_env_or_default("DB_HOSTNAME", "localhost"),
        "port": int(get_env_or_default("DB_PORT", "5432")),
        "username": get_env_or_default("DB_USERNAME", "postgres"),
        "password": db_password,
        "name": get_env_or_default("DB_NAME", "budgets"),
        "schema": get_env_or_default("DB_SCHEMA", "public"),
        "encryption_key": encryption_key,
    }


def generate_db_url(db_config: dict):
    return URL.create(
        drivername=db_config["url_prefix"],
        username=db_config["username"],
        password=db_config["password"],
        host=db_config["hostname"],
        port=db_config["port"],
        database=db_config["name"],
    )



def run_migrations_offline() -> None:
    """Run migrations in 'offline' mode.

    This configures the context with just a URL
    and not an Engine, though an Engine is acceptable
    here as well.  By skipping the Engine creation
    we don't even need a DBAPI to be available.

    Calls to context.execute() here emit the given string to the
    script output.

    """
    url = config.get_main_option("sqlalchemy.url")
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
    )

    with context.begin_transaction():
        context.run_migrations()


def run_migrations_online() -> None:
    """Run migrations in 'online' mode.

    In this scenario we need to create an Engine
    and associate a connection with the context.

    """

    db_config = load_database_config()
    db_url = generate_db_url(db_config)
    configuration = config.get_section(config.config_ini_section)
    configuration["sqlalchemy.url"] = db_url
    connectable = engine_from_config(
        configuration,
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )

    with connectable.connect() as connection:
        context.configure(
            connection=connection,
            target_metadata=target_metadata,
            compare_type=True,
            compare_server_default=True,
        )

        with context.begin_transaction():
            context.run_migrations()


if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()
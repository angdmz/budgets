from sqlalchemy import Column, TIMESTAMP, UUID, func
from sqlalchemy.orm import declarative_base
Base = declarative_base()
metadata = Base.metadata


class BaseModel(Base):
    __abstract__ = True

    created_at = Column(TIMESTAMP(timezone=True),
                        nullable=False,
                        server_default=func.current_timestamp(),)
    updated_at = Column(TIMESTAMP(timezone=True),
                        nullable=False,
                        server_default=func.current_timestamp(),
                        onupdate=func.current_timestamp())


class BaseModelWithID(BaseModel):
    __abstract__ = True
    id = Column(UUID, primary_key=True, server_default=func.uuid_generate_v4())

-- migrate:up tx=true
create table tenants (
                         id uuid primary key not null,
                         name varchar not null,
                         sso_enabled boolean default false,
                         sso_type varchar,
                         sso_provider varchar,
                         oidc_issuer_url varchar,
                         oidc_client_id varchar,
                         oidc_client_secret text,
                         oidc_scopes varchar[] default '{}',
                         saml_entity_id varchar,
                         saml_sso_url varchar,
                         saml_x509_cert text,
                         allowed_email_domains text[] default '{}',
                         created_at timestamp not null default now(),
                         updated_at timestamp not null default now()
);

create table users (
                       id uuid primary key not null,
                       email varchar not null unique,
                       email_verified boolean not null default false,
                       display_name varchar not null,
                       first_name varchar,
                       last_name varchar,
                       timezone varchar,
                       locked_at timestamp,
                       created_at timestamp not null default now(),
                       updated_at timestamp not null default now()
);

create table tenant_users (
                              tenant_id uuid not null references tenants(id) on delete cascade on update cascade,
                              user_id uuid not null references users(id) on delete cascade on update cascade,
                              role varchar not null,
                              created_at timestamp not null default now(),
                              updated_at timestamp not null default now(),
                              primary key (tenant_id, user_id)
);

create table user_identities (
                                 id uuid primary key not null,
                                 tenant_id uuid not null references tenants(id) on delete cascade on update cascade,
                                 user_id uuid not null references users(id) on delete cascade on update cascade,
                                 provider_type varchar not null,
                                 provider varchar not null,
                                 issuer varchar,
                                 subject varchar not null,
                                 created_at timestamp not null default now(),
                                 last_used_at timestamp,
                                 unique (tenant_id, issuer, subject)
);

create table audit_logs (
                            id uuid primary key not null,
                            tenant_id uuid references tenants(id) on delete cascade on update cascade,
                            user_id uuid references users(id) on delete set null on update cascade,
                            action varchar not null,
                            resource_type varchar not null,
                            resource_id uuid,
                            metadata jsonb,
                            ip_address inet,
                            user_agent text,
                            outcome varchar not null,
                            failure_reason varchar,
                            created_at timestamp not null default now()
);

-- migrate:down tx=true
drop table audit_logs;
drop table user_identities;
drop table tenant_users;
drop table users;
drop table tenants;
package constants

const PostgresqlConfContent = `# -----------------------------
# PostgreSQL configuration file
# -----------------------------
#
# DO NOT EDIT THIS FILE
# it will be overwritten on server start
#
#------------------------------------------------------------------------------
# STEAMPIPE CONFIG ADDITIONS
#------------------------------------------------------------------------------

include = 'steampipe.conf'

# settings in this directory will override any steampipe provided values from the
# config file above

include_dir = 'postgresql.conf.d'    # includes files ending in .conf inside this dir
`

const SteampipeConfContent = `
# Custom settings for Steampipe
#
# DO NOT EDIT THIS FILE
# it will be overwritten on server start
#
# Modification or additions to postgres config should be placed in the
# postgresql.conf.d folder with a name alphabetically following this
# file - for example 01-custom-settings.conf
#
# see documentation on this behavior in the postgresql docs:
# https://www.postgresql.org/docs/11/config-setting.html#CONFIG-INCLUDES

autovacuum=off
bgwriter_lru_maxpages=0
effective_cache_size=64kB
fsync=off
full_page_writes=off
maintenance_work_mem=1024kB
password_encryption=scram-sha-256
random_page_cost=0.01
seq_page_cost=0.01

# If the shared buffers are too small then large tables in memory can create
# "no unpinned buffers available" errors.
#
# In that case, set shared_buffers in an overriding config file
# shared_buffers=128kB

# If synchronous_commit=off then the setup process can fail because the
# installation of the foreign server is not committed before the DB shutsdown.
# Steampipe does very few commits in general, so leaving this on will have
# very little impact on performance.
#
# In that case, set synchronous_commit in an overriding config file
# synchronous_commit=off

temp_buffers=800kB
timezone=UTC
track_activities=off
track_counts=off
wal_buffers=32kB
work_mem=64kB
jit=off
max_locks_per_transaction = 2048 

# postgres log collection
log_connections=on
log_disconnections=on
log_min_duration_statement=1000
log_destination=stderr
log_statement=none
log_min_error_statement=error
logging_collector=on
log_filename='database-%Y-%m-%d.log'
log_timezone=UTC
`

# Layered DDD Architecture

On top level there are modules by domain repsponsibility like user, auth, tracking and shared. The layers have to be met for each module. Dependencies between the modules may only occurr on the domain layer. The shared module may not depend on any other module.

A layered ddd architecture is used. The dependencies between the layers is:

1. Presentation layer may depend on domain layer only
2. Domain Layer may not have any dependencies to other layers
3. Infrastructure layer may depend on domain layer only

## Presentation Layer

- REST and Web are the presentation layer.

### Domain Layer

- Services are the domain layer and are used for the use case and logic.
- Repository interfaces are part of the domain layer.
- Domain objects are just simple structs.

### Infrastructure Layer

- Repositories are in the Infrastructure Layer
- There's no need to unit test the in memory repositories.
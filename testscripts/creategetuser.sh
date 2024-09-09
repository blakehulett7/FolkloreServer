echo &&
    echo "Creating user blakehulett..." &&
    echo &&
    curl \
        --header "Content-Type: application/json" \
        --data '{"username":"blakehulett"}' \
        http://localhost:8080/v1/users

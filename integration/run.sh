#!/bin/bash

RESULTS_DIR=/tmp/results

saveResults() {
    cd "${RESULTS_DIR}" || exit
    tar -czf contour.tar.gz ./*
    # mark the done file as a termination notice.
    echo -n "${RESULTS_DIR}/contour.tar.gz" > "${RESULTS_DIR}/done"
}

go test ./... -tags integration > ${RESULTS_DIR}/results.txt
saveResults
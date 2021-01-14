// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGCPRepository(t *testing.T) {
	gcpReg := CommunityGCPBucketRepository
	list, err := gcpReg.List()
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(list), 2)

	_, err = gcpReg.Describe("cluster")
	require.NoError(t, err)

	bin, err := gcpReg.Fetch("cluster", VersionLatest, LinuxAMD64)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(bin), 10)
}

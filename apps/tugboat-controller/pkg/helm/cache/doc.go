/*
Package cache manages in-memory and on-disk caches of charts and chart
repositories.

The `repos` subpackage keeps an in-memory cache of the index files for
repositories represented by `tugboat.engineering.Repository` custom resources.

The `charts` subpackage keeps an on-disk cache of tarballs for charts.

*/
package cache

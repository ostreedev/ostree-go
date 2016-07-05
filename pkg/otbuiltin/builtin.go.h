#include <ostree.h>
#include <string.h>

static void
_ostree_repo_append_modifier_flags(OstreeRepoCommitModifierFlags *flags, int flag) {
  *flags |= flag;
}

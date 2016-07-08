#include <glib.h>
#include <ostree.h>
#include <string.h>

static void
_ostree_repo_append_modifier_flags(OstreeRepoCommitModifierFlags *flags, int flag) {
  *flags |= flag;
}

struct CommitFilterData {
  GHashTable *mode_adds;
  GHashTable *skip_list;
};

typedef struct CommitFilterData CommitFilterData;

static char* _gptr_to_str(gpointer p)
{
    return (char*)p;
}

// The following 3 functions are wrapper functions for macros since CGO can't parse macros
static OstreeRepoFile*
_ostree_repo_file(GFile *file)
{
  return OSTREE_REPO_FILE (file);
}

static guint
_gpointer_to_uint (gpointer ptr)
{
  return GPOINTER_TO_UINT (ptr);
}

static gpointer
_guint_to_pointer (guint u)
{
  return GUINT_TO_POINTER (u);
}

static void
_g_clear_object (volatile GObject **object_ptr)
{
  g_clear_object(object_ptr);
}

static const GVariantType*
_g_variant_type (char *type)
{
  return G_VARIANT_TYPE (type);
}


// These functions are wrappers for variadic functions since CGO can't parse variadic functions
static void
_g_printerr_onearg (char* msg,
                    char* arg)
{
  g_printerr("%s %s\n", msg, arg);
}

static void
_g_set_error_onearg (GError *err,
                     char*  msg,
                     char*  arg)
{
  g_set_error(&err, G_IO_ERROR, G_IO_ERROR_FAILED, "%s %s", msg, arg);
}

static void
_g_variant_builder_add_twoargs (GVariantBuilder*     builder,
                                const gchar   *format_string,
                                char          *arg1,
                                char          *arg2)
{
  g_variant_builder_add(builder, format_string, arg1, arg2);
}

static GHashTable*
_g_hash_table_new_full ()
{
  return g_hash_table_new_full(g_str_hash, g_str_equal, g_free, NULL);
}

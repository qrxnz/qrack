#include <stdio.h>
#include <string.h>

#define PASSWORD "valedor"

int main(int argc, char *argv[]) {
  // Check if the user provided a password as an argument
  if (argc < 2) {
    printf("Usage: %s <password>\n", argv[0]);
    return 1;
  }

  // Compare the provided password with the correct one
  if (strcmp(argv[1], PASSWORD) == 0) {
    printf("Password correct! Access granted.\n");
  } else {
    printf("Incorrect password! Access denied.\n");
  }

  return 0;
}

import crypt
from passlib.hash import sha512_crypt

def generate_sha512_crypt_hash(password, salt=None):
    """
    Generates an SHA-512 crypt-style password hash.

    Args:
        password (str): The plaintext password to hash.
        salt (str, optional): A custom salt to use. If None, a random salt will be generated.

    Returns:
        str: The SHA-512 crypt hash of the password.
    """
    if salt:
        # Prepend '$6$' to indicate SHA-512 crypt algorithm
        # and include the custom salt.
        salt_prefix = f"$6${salt}"
    else:
        # Let crypt generate a random salt for SHA-512 crypt.
        salt_prefix = "$6$"

    hashed_password = crypt.crypt(password, salt_prefix)
    return hashed_password

def hash_password(password: str) -> str:
    hashed = sha512_crypt.hash(password)
    print(hashed)

    is_valid = sha512_crypt.verify(password, hashed)
    if not is_valid:
        raise Exception("Passwords do not match")

    return hashed

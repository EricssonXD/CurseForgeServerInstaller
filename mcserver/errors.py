class McServerError(Exception):
    pass


class UserFacingError(McServerError):
    """Error with a message suitable for CLI output."""


class MissingApiKeyError(UserFacingError):
    pass
